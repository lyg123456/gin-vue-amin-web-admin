package content

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const (
	dashScopeSynthesisURLDefault = "https://dashscope.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis"
	dashScopeTaskURLDefault      = "https://dashscope.aliyuncs.com/api/v1/tasks"
)

// DashScopeVideoService 阿里云 DashScope 异步视频生成（与官方 curl 一致）
type DashScopeVideoService struct{}

func (s *DashScopeVideoService) Enabled() bool {
	return normalizeLLMCredential(global.GVA_CONFIG.DashScopeVideo.ApiKey) != ""
}

func (s *DashScopeVideoService) apiKey() string {
	return normalizeLLMCredential(global.GVA_CONFIG.DashScopeVideo.ApiKey)
}

func (s *DashScopeVideoService) synthesisURL() string {
	u := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.SynthesisURL)
	if u == "" {
		return dashScopeSynthesisURLDefault
	}
	return u
}

func (s *DashScopeVideoService) taskURLBase() string {
	u := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.TaskURL)
	if u == "" {
		return dashScopeTaskURLDefault
	}
	return strings.TrimRight(u, "/")
}

func (s *DashScopeVideoService) httpTimeout() time.Duration {
	sec := global.GVA_CONFIG.DashScopeVideo.Timeout
	if sec <= 0 {
		sec = 600
	}
	return time.Duration(sec) * time.Second
}

func (s *DashScopeVideoService) pollInterval() time.Duration {
	sec := global.GVA_CONFIG.DashScopeVideo.PollIntervalSec
	if sec <= 0 {
		sec = 5
	}
	return time.Duration(sec) * time.Second
}

// mapDashScopeDuration 百炼成片常见 5/10 秒档位
func mapDashScopeDuration(durationSec uint) int {
	d := clampDuration(durationSec)
	if d <= 60 {
		return 5
	}
	return 10
}

func parseSourceImageURLs(sourceImages string) []string {
	var urls []string
	for _, p := range strings.Split(sourceImages, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			urls = append(urls, p)
		}
	}
	return urls
}

func extractVideoScriptBody(script string) string {
	script = strings.TrimSpace(script)
	if script == "" {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(script, "\r\n", "\n"), "\n")
	var bodyLines []string
	inBody := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		lower := strings.ToLower(trim)
		if strings.HasPrefix(trim, "正文") || strings.Contains(lower, "【正文】") {
			if i := strings.IndexAny(trim, "：:"); i >= 0 && i < len(trim)-1 {
				rest := strings.TrimSpace(trim[i+1:])
				if rest != "" {
					bodyLines = append(bodyLines, rest)
				}
			}
			inBody = true
			continue
		}
		if inBody {
			if strings.HasPrefix(trim, "标题") || strings.HasPrefix(trim, "摘要") || strings.HasPrefix(trim, "SEO") {
				break
			}
			bodyLines = append(bodyLines, line)
		}
	}
	if len(bodyLines) > 0 {
		return strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}
	return script
}

func buildVideoPrompt(script, description, promptText string) string {
	body := extractVideoScriptBody(script)
	if body == "" {
		body = strings.TrimSpace(script)
	}
	if body == "" {
		body = strings.TrimSpace(description)
	}
	if body == "" {
		body = strings.TrimSpace(promptText)
	}
	if len([]rune(body)) > 800 {
		body = string([]rune(body)[:800])
	}
	return body
}

func (s *DashScopeVideoService) buildSynthesisBody(script, firstFrameURL, lastFrameURL string, durationSec uint) (map[string]interface{}, string, error) {
	prompt := buildVideoPrompt(script, "", "")
	firstFrameURL, lastFrameURL = NormalizeFrameURLs(firstFrameURL, lastFrameURL, "")
	model := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.Model)
	if model == "" {
		model = "wan2.7-i2v-2026-04-25"
	}
	input := map[string]interface{}{
		"prompt": prompt,
	}
	useI2V := firstFrameURL != "" || lastFrameURL != ""
	if useI2V {
		if err := ValidateI2VFrameURLs(firstFrameURL, lastFrameURL); err != nil {
			return nil, model, err
		}
		firstMedia, err := resolveDashScopeFrameURL(firstFrameURL)
		if err != nil {
			return nil, model, fmt.Errorf("首帧图: %w", err)
		}
		lastMedia, err := resolveDashScopeFrameURL(lastFrameURL)
		if err != nil {
			return nil, model, fmt.Errorf("尾帧图: %w", err)
		}
		input["media"] = []map[string]string{
			{"type": "first_frame", "url": firstMedia},
			{"type": "last_frame", "url": lastMedia},
		}
	} else {
		t2v := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.T2VModel)
		if t2v != "" {
			model = t2v
		}
	}
	resolution := strings.TrimSpace(global.GVA_CONFIG.DashScopeVideo.Resolution)
	if resolution == "" {
		resolution = "720P"
	}
	watermark := global.GVA_CONFIG.DashScopeVideo.Watermark
	body := map[string]interface{}{
		"model": model,
		"input": input,
		"parameters": map[string]interface{}{
			"resolution":     resolution,
			"duration":       mapDashScopeDuration(durationSec),
			"prompt_extend":  false,
			"watermark":      watermark,
		},
	}
	return body, model, nil
}

type dashScopeTaskCreateResp struct {
	Output struct {
		TaskID string `json:"task_id"`
	} `json:"output"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type dashScopeTaskQueryResp struct {
	Output struct {
		TaskStatus string `json:"task_status"`
		VideoURL   string `json:"video_url"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
	} `json:"output"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *DashScopeVideoService) doJSON(ctx context.Context, method, url string, payload []byte) ([]byte, int, error) {
	var body io.Reader
	if len(payload) > 0 {
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey())
	if method == http.MethodPost && strings.Contains(url, "video-synthesis") {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	client := &http.Client{Timeout: s.httpTimeout()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
}

func (s *DashScopeVideoService) createTask(ctx context.Context, script, firstFrameURL, lastFrameURL string, durationSec uint) (string, error) {
	payload, _, err := s.buildSynthesisBody(script, firstFrameURL, lastFrameURL, durationSec)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	b, st, err := s.doJSON(ctx, http.MethodPost, s.synthesisURL(), raw)
	if err != nil {
		return "", err
	}
	if st != http.StatusOK {
		return "", fmt.Errorf("DashScope 提交任务 HTTP %d: %s", st, llmBodySnippet(b, 500))
	}
	var out dashScopeTaskCreateResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", fmt.Errorf("解析提交响应失败: %w", err)
	}
	if out.Code != "" && out.Code != "Success" {
		return "", fmt.Errorf("DashScope 提交失败: %s %s", out.Code, out.Message)
	}
	taskID := strings.TrimSpace(out.Output.TaskID)
	if taskID == "" {
		return "", fmt.Errorf("DashScope 未返回 task_id: %s", llmBodySnippet(b, 300))
	}
	return taskID, nil
}

func (s *DashScopeVideoService) queryTask(ctx context.Context, taskID string) (status, videoURL string, err error) {
	u := s.taskURLBase() + "/" + taskID
	b, st, err := s.doJSON(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}
	if st != http.StatusOK {
		return "", "", fmt.Errorf("DashScope 查询任务 HTTP %d: %s", st, llmBodySnippet(b, 500))
	}
	var out dashScopeTaskQueryResp
	if err := json.Unmarshal(b, &out); err != nil {
		return "", "", err
	}
	if out.Code != "" && out.Code != "Success" && out.Output.TaskStatus == "" {
		return "", "", fmt.Errorf("DashScope 查询失败: %s %s", out.Code, out.Message)
	}
	status = strings.ToUpper(strings.TrimSpace(out.Output.TaskStatus))
	videoURL = strings.TrimSpace(out.Output.VideoURL)
	if videoURL == "" && len(out.Output.Results) > 0 {
		videoURL = strings.TrimSpace(out.Output.Results[0].URL)
	}
	return status, videoURL, nil
}

// SubmitGeneration 提交并轮询直至成功/失败/超时（i2v 需 firstFrameURL + lastFrameURL）
func (s *DashScopeVideoService) SubmitGeneration(script, firstFrameURL, lastFrameURL string, durationSec uint) (taskID, videoURL string, err error) {
	if !s.Enabled() {
		return "", "", errors.New("未配置 dashscope-video.api-key（阿里云 DASHSCOPE_API_KEY）")
	}
	firstFrameURL, lastFrameURL = NormalizeFrameURLs(firstFrameURL, lastFrameURL, "")
	// 默认 i2v 模型需首尾两帧；无图时走 t2v 模型
	if firstFrameURL != "" || lastFrameURL != "" {
		if err := ValidateI2VFrameURLs(firstFrameURL, lastFrameURL); err != nil {
			return "", "", err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.httpTimeout())
	defer cancel()

	taskID, err = s.createTask(ctx, script, firstFrameURL, lastFrameURL, durationSec)
	if err != nil {
		return "", "", err
	}

	ticker := time.NewTicker(s.pollInterval())
	defer ticker.Stop()
	for {
		status, url, qerr := s.queryTask(ctx, taskID)
		if qerr != nil {
			return taskID, "", qerr
		}
		switch status {
		case "SUCCEEDED", "SUCCESS":
			if url == "" {
				return taskID, "", errors.New("任务成功但未返回 video_url，请稍后在控制台查看")
			}
			return taskID, url, nil
		case "FAILED", "CANCELED", "CANCELLED":
			return taskID, "", fmt.Errorf("DashScope 视频任务失败: %s", status)
		case "PENDING", "RUNNING", "":
			// continue poll
		default:
			// unknown state, keep polling
		}
		select {
		case <-ctx.Done():
			return taskID, "", fmt.Errorf("等待 DashScope 成片超时，task_id=%s（可稍后凭 task_id 查询）", taskID)
		case <-ticker.C:
		}
	}
}
