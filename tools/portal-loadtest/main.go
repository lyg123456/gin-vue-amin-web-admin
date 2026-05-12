package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

// Config 从 YAML 加载；路径见 config.example.yaml
type Config struct {
	Target         string   `yaml:"target"`
	Timezone       string   `yaml:"timezone"`
	Windows        []Window `yaml:"windows"`
	QPSMin         int      `yaml:"qps_min"`
	QPSMax         int      `yaml:"qps_max"`
	Concurrency    int      `yaml:"concurrency"`
	RequestTimeout string   `yaml:"request_timeout"`
	ExtraPaths     []string `yaml:"extra_paths"`
	// IdleInterval 非活跃时段：两次「单请求心跳」之间的间隔，默认 1h
	IdleInterval string `yaml:"idle_interval"`
}

type Window struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	once := flag.Bool("once", false, "ignore time windows: run ~30s load then exit (smoke test)")
	flag.Parse()

	raw, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		log.Fatalf("yaml: %v", err)
	}
	if err := validate(&cfg); err != nil {
		log.Fatalf("config: %v", err)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("timezone: %v", err)
	}

	timeout, err := time.ParseDuration(strings.TrimSpace(cfg.RequestTimeout))
	if err != nil || timeout <= 0 {
		timeout = 8 * time.Second
	}

	base, err := url.Parse(strings.TrimSpace(cfg.Target))
	if err != nil || base.Scheme == "" || base.Host == "" {
		log.Fatalf("invalid target URL")
	}

	paths := []string{"/"}
	if len(cfg.ExtraPaths) > 0 {
		paths = cfg.ExtraPaths
	}
	urls := make([]string, 0, len(paths))
	for _, p := range paths {
		ref, err := url.Parse(p)
		if err != nil {
			continue
		}
		u := base.ResolveReference(ref)
		urls = append(urls, u.String())
	}

	client := newHTTPClient(cfg.Concurrency, timeout)

	idleEvery, err := time.ParseDuration(strings.TrimSpace(cfg.IdleInterval))
	if err != nil || idleEvery < time.Minute {
		idleEvery = time.Hour
	}

	log.Printf("portal-loadtest: target=%s urls=%v qps=[%d,%d] concurrency=%d tz=%s idle_ping=%v",
		base.String(), urls, cfg.QPSMin, cfg.QPSMax, cfg.Concurrency, cfg.Timezone, idleEvery)
	if *once {
		runBurst(context.Background(), client, urls, &cfg, 30*time.Second)
		return
	}

	for {
		now := time.Now().In(loc)
		if !inAnyWindow(now, loc, cfg.Windows) {
			nxt := nextWindowStart(now, loc, cfg.Windows)
			if nxt.IsZero() {
				log.Fatal("no windows configured")
			}
			doIdlePing(context.Background(), client, urls)
			sleep := idleEvery
			if until := time.Until(nxt); until < sleep {
				sleep = until
			}
			if sleep < time.Second {
				sleep = time.Second
			}
			log.Printf("outside active windows: idle ping done, sleep %v (next window at %s)",
				sleep.Round(time.Second), nxt.Format(time.RFC3339))
			time.Sleep(sleep)
			continue
		}

		// 每秒换一档随机 QPS
		secStart := time.Now()
		qps := randInt(cfg.QPSMin, cfg.QPSMax+1)
		runOneSecond(context.Background(), client, urls, cfg.Concurrency, qps)
		elapsed := time.Since(secStart)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
	}
}

func validate(c *Config) error {
	if strings.TrimSpace(c.Target) == "" {
		return fmt.Errorf("target is required")
	}
	if c.Timezone == "" {
		c.Timezone = "Asia/Shanghai"
	}
	if len(c.Windows) == 0 {
		return fmt.Errorf("windows is required")
	}
	for i, w := range c.Windows {
		if _, err := parseHM(w.Start); err != nil {
			return fmt.Errorf("windows[%d].start: %w", i, err)
		}
		if _, err := parseHM(w.End); err != nil {
			return fmt.Errorf("windows[%d].end: %w", i, err)
		}
	}
	if c.QPSMin <= 0 || c.QPSMax < c.QPSMin {
		return fmt.Errorf("invalid qps_min / qps_max")
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 1000
	}
	if c.RequestTimeout == "" {
		c.RequestTimeout = "8s"
	}
	if strings.TrimSpace(c.IdleInterval) == "" {
		c.IdleInterval = "1h"
	}
	return nil
}

func newHTTPClient(maxIdle int, timeout time.Duration) *http.Client {
	if maxIdle < 1 {
		maxIdle = 1
	}
	tr := &http.Transport{
		MaxIdleConns:        maxIdle,
		MaxIdleConnsPerHost: maxIdle,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	return &http.Client{Transport: tr, Timeout: timeout}
}

func doIdlePing(ctx context.Context, client *http.Client, urls []string) {
	if len(urls) == 0 {
		return
	}
	u := urls[randInt(0, len(urls))]
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		log.Printf("idle ping: build request: %v", err)
		return
	}
	req.Header.Set("User-Agent", "portal-loadtest/1.0-idle (+https://github.com/flipped-aurora/gin-vue-admin)")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("idle ping: %s err=%v", u, err)
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	log.Printf("idle ping: %s status=%d", u, resp.StatusCode)
}

func runBurst(ctx context.Context, client *http.Client, urls []string, cfg *Config, dur time.Duration) {
	deadline := time.Now().Add(dur)
	log.Printf("-once mode: running for %v", dur)
	for time.Now().Before(deadline) {
		qps := randInt(cfg.QPSMin, cfg.QPSMax+1)
		sub, cancel := context.WithTimeout(ctx, time.Second)
		runOneSecond(sub, client, urls, cfg.Concurrency, qps)
		cancel()
	}
	log.Printf("-once mode: done")
}

func runOneSecond(ctx context.Context, client *http.Client, urls []string, workers int, qps int) {
	var okCount, errCount int64
	limit := rate.Limit(float64(qps))
	burst := qps / 10
	if burst < 50 {
		burst = 50
	}
	if burst > qps {
		burst = qps
	}
	if burst > 2000 {
		burst = 2000
	}
	lim := rate.NewLimiter(limit, burst)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if err := lim.Wait(ctx); err != nil {
					return
				}
				u := urls[randInt(0, len(urls))]
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					continue
				}
				req.Header.Set("User-Agent", "portal-loadtest/1.0 (+https://github.com/flipped-aurora/gin-vue-admin)")
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					atomic.AddInt64(&okCount, 1)
				} else {
					atomic.AddInt64(&errCount, 1)
				}
			}
		}()
	}
	wg.Wait()
	if okCount+errCount > 0 {
		log.Printf("1s tick: qps_target=%d ok=%d err_or_4xx=%d", qps, okCount, errCount)
	}
}

func minuteOfDay(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

func parseHM(s string) (int, error) {
	var h, m int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d:%d", &h, &m)
	if err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	return h*60 + m, nil
}

func inAnyWindow(t time.Time, loc *time.Location, wins []Window) bool {
	t = t.In(loc)
	mod := minuteOfDay(t)
	for _, w := range wins {
		sm, err1 := parseHM(w.Start)
		em, err2 := parseHM(w.End)
		if err1 != nil || err2 != nil {
			continue
		}
		if mod >= sm && mod <= em {
			return true
		}
	}
	return false
}

// nextWindowStart 返回严格晚于 now 的下一个时段「开始」时刻（整点秒落在 start 分钟起点）
func nextWindowStart(now time.Time, loc *time.Location, wins []Window) time.Time {
	now = now.In(loc)
	var best time.Time
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	for add := 0; add <= 1; add++ {
		d := day.AddDate(0, 0, add)
		for _, w := range wins {
			sm, err := parseHM(w.Start)
			if err != nil {
				continue
			}
			h, m := sm/60, sm%60
			cand := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, loc)
			if cand.After(now) && (best.IsZero() || cand.Before(best)) {
				best = cand
			}
		}
	}
	return best
}

func randInt(min, maxExclusive int) int {
	if maxExclusive <= min {
		return min
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	v := binary.BigEndian.Uint64(b[:])
	return min + int(v%uint64(maxExclusive-min))
}
