package content

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const dashScopeFrameMaxBytes = 10 << 20 // 10MB

var dashScopeFrameMIME = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".bmp":  "image/bmp",
	".webp": "image/webp",
}

// resolveDashScopeFrameURL 将首尾帧转为 DashScope 可识别的公网 URL 或 data:image/...;base64,...
func resolveDashScopeFrameURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("图片地址为空")
	}
	if strings.HasPrefix(raw, "data:image/") {
		return raw, nil
	}

	if publicURL, ok := tryPublicFrameURL(raw); ok {
		return publicURL, nil
	}

	if filePath, ok := localFrameFilePath(raw); ok {
		return imageFileToDataURL(filePath)
	}

	return "", fmt.Errorf(
		"首尾帧须为公网 HTTPS 图片（JPG/PNG/BMP/WEBP），或本地上传文件；当前地址 DashScope 无法访问：%s。可在 config.yaml 配置 dashscope-video.public-file-base-url",
		truncateForErr(raw, 120),
	)
}

func tryPublicFrameURL(raw string) (string, bool) {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", false
		}
		host := strings.ToLower(u.Hostname())
		if host == "localhost" || host == "127.0.0.1" || isPrivateHost(host) {
			return "", false
		}
		return raw, true
	}

	base := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.PublicFileBaseURL)
	if base == "" {
		return "", false
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return "", false
	}
	path := strings.TrimLeft(strings.TrimSpace(raw), "/")
	return strings.TrimRight(base, "/") + "/" + path, true
}

func isPrivateHost(host string) bool {
	if strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") {
		return true
	}
	if strings.HasPrefix(host, "172.") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}

func localFrameFilePath(raw string) (string, bool) {
	path := raw
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", false
		}
		path = strings.TrimPrefix(u.Path, "/")
	}
	path = strings.TrimPrefix(strings.TrimSpace(path), "/")
	path = filepath.FromSlash(path)

	candidates := []string{
		path,
		filepath.Join(global.GVA_CONFIG.Local.StorePath, filepath.Base(path)),
		filepath.Join(global.GVA_CONFIG.Local.Path, filepath.Base(path)),
	}
	seen := map[string]struct{}{}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

func imageFileToDataURL(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	mime, ok := dashScopeFrameMIME[ext]
	if !ok {
		return "", fmt.Errorf("不支持的图片格式 %s，请使用 JPG/PNG/BMP/WEBP", ext)
	}
	st, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("读取图片失败: %w", err)
	}
	if st.Size() > dashScopeFrameMaxBytes {
		return "", fmt.Errorf("图片过大（>%dMB），请压缩后重试", dashScopeFrameMaxBytes>>20)
	}
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取图片失败: %w", err)
	}
	if len(b) == 0 {
		return "", fmt.Errorf("图片文件为空")
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

func truncateForErr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
