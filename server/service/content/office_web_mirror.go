package content

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type OfficeWebStyleService struct{}

type mirrorJob struct {
	URL   string
	Depth int
}

var (
	hrefSrcRe = regexp.MustCompile(`(?i)(?:href|src)\s*=\s*["']([^"'#][^"']*)["']`)
	skipExt   = map[string]bool{
		".pdf": true, ".zip": true, ".rar": true, ".exe": true, ".dmg": true,
		".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true,
		".mp4": true, ".mp3": true, ".avi": true, ".mov": true,
	}
)

func (s *OfficeWebStyleService) mirrorMaxPages() int {
	n := global.GVA_CONFIG.OfficeTools.MirrorMaxPages
	if n <= 0 {
		return 50
	}
	return n
}

func (s *OfficeWebStyleService) mirrorMaxDepth() int {
	n := global.GVA_CONFIG.OfficeTools.MirrorMaxDepth
	if n <= 0 {
		return 3
	}
	return n
}

// DownloadWebsiteZIP 爬取同域前端页面（HTML + 同域静态资源）打包 ZIP
func (s *OfficeWebStyleService) DownloadWebsiteZIP(startURL string, optPages, optDepth int) ([]byte, string, error) {
	start, err := validateOfficeWebURL(startURL)
	if err != nil {
		return nil, "", err
	}
	baseHost := start.Host
	maxPages := s.mirrorMaxPages()
	maxDepth := s.mirrorMaxDepth()
	if optPages > 0 {
		maxPages = optPages
	}
	if optDepth > 0 {
		maxDepth = optDepth
	}

	type fileEntry struct {
		zipPath string
		data    []byte
	}
	var (
		files   []fileEntry
		filesMu sync.Mutex
		visited = map[string]bool{}
		visitMu sync.Mutex
		total   int64
	)

	addFile := func(zipPath string, data []byte) bool {
		filesMu.Lock()
		defer filesMu.Unlock()
		total += int64(len(data))
		if total > officeWebMirrorMaxBytes() {
			return false
		}
		files = append(files, fileEntry{zipPath: zipPath, data: data})
		return true
	}

	queue := []mirrorJob{{URL: start.String(), Depth: 0}}
	pageCount := 0

	for len(queue) > 0 && pageCount < maxPages {
		job := queue[0]
		queue = queue[1:]

		u, err := validateOfficeWebURL(job.URL)
		if err != nil || u.Host != baseHost {
			continue
		}
		norm := u.String()
		visitMu.Lock()
		if visited[norm] {
			visitMu.Unlock()
			continue
		}
		visited[norm] = true
		visitMu.Unlock()

		body, contentType, err := fetchOfficeWebRaw(u.String())
		if err != nil {
			continue
		}
		if total+int64(len(body)) > officeWebMirrorMaxBytes() {
			break
		}

		isHTML := strings.Contains(contentType, "text/html") || strings.HasSuffix(strings.ToLower(u.Path), ".html") || strings.HasSuffix(strings.ToLower(u.Path), ".htm") || u.Path == "" || strings.HasSuffix(u.Path, "/")
		if !isHTML && !isMirrorAssetPath(u.Path) {
			continue
		}

		zipPath := mirrorURLToZipPath(u, isHTML)
		if !addFile(zipPath, body) {
			break
		}
		if isHTML {
			pageCount++
		}

		if job.Depth >= maxDepth || !isHTML {
			continue
		}

		links := extractMirrorLinks(string(body), u)
		for _, link := range links {
			nu, err := validateOfficeWebURL(link)
			if err != nil || nu.Host != baseHost {
				continue
			}
			nlink := nu.String()
			if isMirrorAssetPath(nu.Path) {
				visitMu.Lock()
				seen := visited[nlink]
				if !seen {
					visited[nlink] = true
				}
				visitMu.Unlock()
				if seen {
					continue
				}
				abuf, _, err := fetchOfficeWebRaw(nlink)
				if err != nil {
					continue
				}
				if !addFile(mirrorURLToZipPath(nu, false), abuf) {
					break
				}
				continue
			}
			if shouldSkipMirrorPath(nu.Path) {
				continue
			}
			visitMu.Lock()
			seen := visited[nlink]
			visitMu.Unlock()
			if seen {
				continue
			}
			queue = append(queue, mirrorJob{URL: nlink, Depth: job.Depth + 1})
		}
	}

	if pageCount == 0 {
		return nil, "", fmt.Errorf("未下载到任何页面，请确认网址可访问且为静态或可抓取的 HTML 站点（纯前端渲染站点可能无效）")
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	readme := fmt.Sprintf("网站前端页面打包\n起始网址: %s\n下载时间: %s\n页面数(HTML): %d\n文件数: %d\n说明: 仅爬取同域链接，深度≤%d，最多%d页；临时数据服务器保留约24小时后清理。\n",
		start.String(), time.Now().Format(time.RFC3339), pageCount, len(files), maxDepth, maxPages)
	_ = writeZipFile(zw, "README.txt", []byte(readme))
	for _, f := range files {
		if err := writeZipFile(zw, f.zipPath, f.data); err != nil {
			return nil, "", err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}
	name := "website-pages-" + sanitizeMirrorFilename(start.Hostname()) + "-" + time.Now().Format("20060102-150405") + ".zip"
	return buf.Bytes(), name, nil
}

func officeWebMirrorMaxBytes() int64 {
	mb := global.GVA_CONFIG.OfficeTools.MirrorMaxMB
	if mb <= 0 {
		mb = 30
	}
	return int64(mb) * 1024 * 1024
}

func fetchOfficeWebRaw(pageURL string) ([]byte, string, error) {
	u, err := validateOfficeWebURL(pageURL)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GVA-OfficeTools-Mirror/1.0)")
	req.Header.Set("Accept", "*/*")
	resp, err := officeWebClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, "", err
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = mime.TypeByExtension(path.Ext(u.Path))
	}
	return body, ct, nil
}

func normalizeMirrorURL(raw string, base *url.URL) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "javascript:") || strings.HasPrefix(raw, "mailto:") || strings.HasPrefix(raw, "tel:") || strings.HasPrefix(raw, "data:") {
		return ""
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	abs.Fragment = ""
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

func extractMirrorLinks(html string, base *url.URL) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range hrefSrcRe.FindAllStringSubmatch(html, -1) {
		if len(m) < 2 {
			continue
		}
		n := normalizeMirrorURL(m[1], base)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

func shouldSkipMirrorPath(p string) bool {
	ext := strings.ToLower(path.Ext(p))
	return skipExt[ext]
}

func isMirrorAssetPath(p string) bool {
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".css", ".js", ".mjs", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map":
		return true
	default:
		return false
	}
}

func mirrorURLToZipPath(u *url.URL, forceHTML bool) string {
	hostDir := sanitizeMirrorFilename(u.Hostname())
	p := u.Path
	if p == "" || p == "/" {
		return filepath.Join(hostDir, "index.html")
	}
	p = strings.TrimPrefix(p, "/")
	if strings.HasSuffix(p, "/") {
		return filepath.Join(hostDir, filepath.FromSlash(p), "index.html")
	}
	ext := strings.ToLower(path.Ext(p))
	if forceHTML && ext != ".html" && ext != ".htm" {
		// /about -> about.html 便于本地打开
		if ext == "" {
			return filepath.Join(hostDir, filepath.FromSlash(p)+".html")
		}
	}
	return filepath.Join(hostDir, filepath.FromSlash(p))
}

func sanitizeMirrorFilename(s string) string {
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return r
		}
		return '_'
	}, s)
	if s == "" {
		return "site"
	}
	return s
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(filepath.ToSlash(name))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
