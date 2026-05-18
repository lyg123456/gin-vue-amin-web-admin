package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// VolcArkShortVideoService 火山大模型生成短视频脚本（复用 volc-ark chat 配置）
type VolcArkShortVideoService struct{}

// ShortVideoScriptReq 生成脚本入参
type ShortVideoScriptReq struct {
	Title       string
	Description string
	PromptText  string
	DurationSec uint
}

func (s *VolcArkShortVideoService) ensureChatConfig() error {
	if !volcArkArticleService.Enabled() {
		return errors.New("请先在 server/config.yaml 配置 volc-ark.api-key（用于生成短视频脚本）")
	}
	return nil
}

func clampDuration(sec uint) uint {
	if sec < 30 {
		return 30
	}
	if sec > 120 {
		return 120
	}
	return sec
}

func (s *VolcArkShortVideoService) buildScriptPrompt(req ShortVideoScriptReq) string {
	dur := clampDuration(req.DurationSec)
	var b strings.Builder
	b.WriteString("你是资深短视频编导。请根据以下信息输出一份可直接拍摄的短视频脚本。\n\n")
	b.WriteString(fmt.Sprintf("【标题】%s\n", strings.TrimSpace(req.Title)))
	if t := strings.TrimSpace(req.PromptText); t != "" {
		b.WriteString(fmt.Sprintf("【创意文字】%s\n", t))
	}
	if d := strings.TrimSpace(req.Description); d != "" {
		b.WriteString(fmt.Sprintf("【场景/产品描述】%s\n", d))
	}
	b.WriteString(fmt.Sprintf("【目标时长】约 %d 秒（口播+画面，可略浮动）\n\n", dur))
	b.WriteString(`【输出格式】严格按以下标签分段（标签后只写该段内容，不要重复标签名）：
标题：（一行，可不与上面重复）
摘要：（50字内口播摘要）
SEO描述：（80字内，用于发布简介）
SEO关键词：（逗号分隔，3-8个）
正文：（分镜脚本正文：按时间轴 0-5s、5-15s… 写画面+旁白+字幕要点；不要写「标题」「SEO」等元信息）
`)
	return b.String()
}

func (s *VolcArkShortVideoService) GenerateScript(req ShortVideoScriptReq) (string, error) {
	if strings.TrimSpace(req.Title) == "" {
		return "", errors.New("标题不能为空")
	}
	if err := s.ensureChatConfig(); err != nil {
		return "", err
	}
	req.DurationSec = clampDuration(req.DurationSec)
	prompt := s.buildScriptPrompt(req)
	st, body, err := volcArkArticleService.postChat(prompt, 0.7)
	if err != nil {
		return "", err
	}
	if st != http.StatusOK {
		return "", fmt.Errorf("火山脚本接口 HTTP %d %s", st, llmBodySnippet(body, 400))
	}
	text, err := parseOpenAIChatCompletion(body, st, "火山方舟")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// VolcArkVideoGenerateService 短视频成片生成（预留火山视频 API）
type VolcArkVideoGenerateService struct{}

func (s *VolcArkVideoGenerateService) Enabled() bool {
	w := global.GVA_CONFIG.VolcArkVideo
	return normalizeLLMCredential(w.ApiKey) != "" && strings.TrimSpace(w.BaseURL) != ""
}

func (s *VolcArkVideoGenerateService) httpTimeout() time.Duration {
	sec := global.GVA_CONFIG.VolcArkVideo.Timeout
	if sec <= 0 {
		sec = 300
	}
	return time.Duration(sec) * time.Second
}

// SubmitGeneration 提交成片任务；未配置视频 AK 时返回明确错误
func (s *VolcArkVideoGenerateService) SubmitGeneration(script, sourceImages string, durationSec uint) (taskID string, videoURL string, err error) {
	if !s.Enabled() {
		return "", "", errors.New("未配置 volc-ark-video.api-key 与 base-url，暂无法自动生成成片。请先在 config.yaml 填写火山短视频 API，或手动上传视频 URL")
	}
	_ = script
	_ = sourceImages
	_ = durationSec
	// 预留：对接火山视频生成 OpenAPI 后在此 POST 并解析 task_id / video_url
	base := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.VolcArkVideo.BaseURL), "/")
	payload := map[string]interface{}{
		"model": global.GVA_CONFIG.VolcArkVideo.Model,
		"script": script,
	}
	raw, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest(http.MethodPost, base, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+normalizeLLMCredential(global.GVA_CONFIG.VolcArkVideo.ApiKey))
	client := &http.Client{Timeout: s.httpTimeout()}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("火山视频接口 HTTP %d %s", resp.StatusCode, llmBodySnippet(body, 400))
	}
	// 通用解析占位
	var out struct {
		TaskID   string `json:"task_id"`
		VideoURL string `json:"video_url"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("解析视频任务响应失败: %w", err)
	}
	return out.TaskID, out.VideoURL, nil
}
