package content

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const mediaDownloadMaxBytes = 120 << 20 // 120MB

var allowedMediaHostSuffixes = []string{
	"xhscdn.com",
	"xiaohongshu.com",
	"douyin.com",
	"douyinvod.com",
	"amemv.com",
	"snssdk.com",
	"bytecdn.cn",
	"byteimg.com",
	"ixigua.com",
	"weixin.qq.com",
	"qq.com",
}

// MediaDownloadReq 服务端代下（携带 Referer/Cookie，绕过 CDN 403）
type MediaDownloadReq struct {
	URL      string `json:"url"`
	Cookie   string `json:"cookie"`
	Platform string `json:"platform"` // xhs | douyin
	Title    string `json:"title"`
}

func mediaReferer(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "douyin":
		return "https://www.douyin.com/"
	case "xhs", "xiaohongshu":
		return "https://www.xiaohongshu.com/"
	default:
		return "https://www.xiaohongshu.com/"
	}
}

func isAllowedMediaURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	for _, suf := range allowedMediaHostSuffixes {
		if host == suf || strings.HasSuffix(host, "."+suf) {
			return true
		}
	}
	return false
}

func sanitizeDownloadFilename(title, noteID string) string {
	base := strings.TrimSpace(title)
	if base == "" {
		base = noteID
	}
	if base == "" {
		base = "video"
	}
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	base = re.ReplaceAllString(base, "_")
	if len([]rune(base)) > 60 {
		base = string([]rune(base)[:60])
	}
	return base + ".mp4"
}

// StreamMediaDownload 代拉媒体并写入 w
func StreamMediaDownload(req MediaDownloadReq, w http.ResponseWriter) error {
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return fmt.Errorf("缺少下载地址")
	}
	if !isAllowedMediaURL(rawURL) {
		return fmt.Errorf("不允许的下载域名")
	}
	ref := mediaReferer(req.Platform)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("User-Agent", xhsDefaultUA)
	httpReq.Header.Set("Referer", ref)
	httpReq.Header.Set("Origin", strings.TrimSuffix(ref, "/"))
	if c := strings.TrimSpace(req.Cookie); c != "" {
		httpReq.Header.Set("Cookie", normalizeXhsCookie(c))
	}
	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("拉取视频失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("CDN 返回 HTTP %d（直链可能已过期，请重新抓取）", resp.StatusCode)
	}
	ctype := resp.Header.Get("Content-Type")
	if ctype == "" {
		ctype = "video/mp4"
	}
	filename := sanitizeDownloadFilename(req.Title, "")
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if i := strings.Index(cd, "filename="); i >= 0 {
			fn := strings.Trim(cd[i+9:], `"' `)
			if fn != "" {
				filename = fn
			}
		}
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".mp4") {
		ext := path.Ext(path.Base(rawURL))
		if ext == "" {
			ext = ".mp4"
		}
		if !strings.HasSuffix(strings.ToLower(filename), strings.ToLower(ext)) {
			filename = strings.TrimSuffix(filename, ".mp4") + ext
		}
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, err = io.Copy(w, io.LimitReader(resp.Body, mediaDownloadMaxBytes))
	return err
}
