package content

import (
	"crypto/rand"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OfficeSpeedService struct{}

type SpeedPingResult struct {
	ServerTime int64  `json:"serverTime"`
	ServerMs   int64  `json:"serverMs"`
	Message    string `json:"message"`
}

type SpeedInfoResult struct {
	ClientIP   string `json:"clientIP"`
	ServerTime int64  `json:"serverTime"`
	Node       string `json:"node"`
}

func (s *OfficeSpeedService) Ping() SpeedPingResult {
	now := time.Now()
	return SpeedPingResult{
		ServerTime: now.Unix(),
		ServerMs:   now.UnixMilli(),
		Message:    "pong",
	}
}

func (s *OfficeSpeedService) ClientInfo(c *gin.Context) SpeedInfoResult {
	ip := c.ClientIP()
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ip = xff
	}
	if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil && ip == "" {
		ip = host
	}
	return SpeedInfoResult{
		ClientIP:   ip,
		ServerTime: time.Now().UnixMilli(),
		Node:       "office-speed",
	}
}

// DownloadPayload 生成指定字节数的测速数据（禁用缓存）
func (s *OfficeSpeedService) DownloadPayload(sizeBytes int) []byte {
	if sizeBytes <= 0 {
		sizeBytes = 1024 * 1024
	}
	const max = 20 * 1024 * 1024
	if sizeBytes > max {
		sizeBytes = max
	}
	buf := make([]byte, sizeBytes)
	_, _ = rand.Read(buf)
	return buf
}

func (s *OfficeSpeedService) ParseSizeMB(c *gin.Context, def int) int {
	mb, _ := strconv.Atoi(c.DefaultQuery("mb", strconv.Itoa(def)))
	if mb <= 0 {
		mb = def
	}
	if mb > 20 {
		mb = 20
	}
	return mb
}

func (s *OfficeSpeedService) WriteDownload(c *gin.Context, mb int) {
	data := s.DownloadPayload(mb * 1024 * 1024)
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

type SpeedUploadResult struct {
	Bytes int64 `json:"bytes"`
}

func (s *OfficeSpeedService) UploadEcho(c *gin.Context) (SpeedUploadResult, error) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 25*1024*1024))
	if err != nil {
		return SpeedUploadResult{}, err
	}
	return SpeedUploadResult{Bytes: int64(len(body))}, nil
}
