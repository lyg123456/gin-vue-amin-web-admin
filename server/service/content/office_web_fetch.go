package content

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const officeWebMaxBody = 3 << 20 // 3MB

var officeWebClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("重定向过多")
		}
		if _, err := validateOfficeWebURL(req.URL.String()); err != nil {
			return err
		}
		return nil
	},
}

func validateOfficeWebURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("请输入网址")
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("网址无效: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("仅支持 http/https")
	}
	if u.Host == "" {
		return nil, errors.New("网址缺少域名")
	}
	host := u.Hostname()
	if isBlockedOfficeHost(host) {
		return nil, errors.New("不允许访问内网或本地地址")
	}
	ips, err := net.LookupIP(host)
	if err == nil {
		for _, ip := range ips {
			if isBlockedOfficeIP(ip) {
				return nil, errors.New("不允许访问内网地址")
			}
		}
	}
	return u, nil
}

func isBlockedOfficeHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "localhost" || strings.HasSuffix(h, ".local") {
		return true
	}
	return isBlockedOfficeIP(net.ParseIP(h))
}

func isBlockedOfficeIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
		return true
	}
	return false
}

func fetchOfficeWebPage(pageURL string) (string, string, error) {
	u, err := validateOfficeWebURL(pageURL)
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GVA-OfficeTools/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := officeWebClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("抓取失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("上游 HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, officeWebMaxBody))
	if err != nil {
		return "", "", err
	}
	ct := resp.Header.Get("Content-Type")
	html := string(body)
	if strings.Contains(ct, "charset=gbk") || strings.Contains(ct, "charset=GBK") {
		// 保持原样，正则仍可匹配 ASCII 片段
	}
	return html, u.String(), nil
}

var hexColorRe = regexp.MustCompile(`#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\b`)

func extractPageStyle(html, finalURL string) map[string]string {
	title := extractMeta(html, `<title[^>]*>([^<]*)</title>`)
	if title == "" {
		title = "网站风格模板"
	}
	themeColor := extractMeta(html, `theme-color["'\s]+content=["']([^"']+)`)
	if themeColor == "" {
		themeColor = extractMeta(html, `property=["']og:color["']\s+content=["']([^"']+)`)
	}
	fontFamily := extractMeta(html, `font-family\s*:\s*([^;}"']+)`)
	if fontFamily == "" {
		fontFamily = "system-ui, -apple-system, sans-serif"
	}
	fontFamily = strings.TrimSpace(strings.Split(fontFamily, ",")[0])
	fontFamily = strings.Trim(fontFamily, `"'`)

	colors := collectColors(html)
	primary := "#2563eb"
	secondary := "#64748b"
	accent := "#f59e0b"
	if themeColor != "" && strings.HasPrefix(themeColor, "#") {
		primary = themeColor
	}
	if len(colors) > 0 {
		primary = colors[0]
	}
	if len(colors) > 1 {
		secondary = colors[1]
	}
	if len(colors) > 2 {
		accent = colors[2]
	}
	siteName := extractMeta(html, `property=["']og:site_name["']\s+content=["']([^"']+)`)
	if siteName == "" {
		if u, err := url.Parse(finalURL); err == nil {
			siteName = u.Hostname()
		}
	}
	return map[string]string{
		"title":       title,
		"siteName":    siteName,
		"primary":     primary,
		"secondary":   secondary,
		"accent":      accent,
		"fontFamily":  fontFamily,
		"sourceUrl":   finalURL,
	}
}

func extractMeta(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	if m := re.FindStringSubmatch(html); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func collectColors(html string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range hexColorRe.FindAllString(html, -1) {
		c := strings.ToLower(m)
		if c == "#fff" || c == "#ffffff" || c == "#000" || c == "#000000" {
			continue
		}
		if len(c) == 4 {
			c = expandShortHex(c)
		}
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func expandShortHex(c string) string {
	if len(c) != 4 {
		return c
	}
	return "#" + string([]byte{c[1], c[1], c[2], c[2], c[3], c[3]})
}
