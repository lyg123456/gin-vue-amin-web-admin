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
	"go.uber.org/zap"
)

const (
	volcArkChatURLDefault = "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
)

// volcArkEndpointHint 404 InvalidEndpointOrModel 时的说明（与官方控制台「推理接入点」一致）
const volcArkEndpointHint = `方舟要求请求体里的「model」为控制台「推理接入点」的 Endpoint ID（通常以 ep- 开头），不是商品名「Doubao-lite-4k」。请打开 https://console.volcengine.com/ark 在线推理 → 推理接入点，复制已启用接入点的 ID 填入 volc-ark.model，并确认 chat-url 地域与该接入点一致。`

func volcArkWrapHTTPError(status int, body []byte) error {
	sn := llmBodySnippet(body, 800)
	msg := fmt.Sprintf("火山方舟接口错误: HTTP %d %s", status, sn)
	if status == http.StatusNotFound && bytes.Contains(body, []byte("InvalidEndpointOrModel")) {
		msg += " " + volcArkEndpointHint
	}
	return errors.New(msg)
}

// VolcArkArticleService 火山方舟（豆包）chat/completions
type VolcArkArticleService struct{}

func (s *VolcArkArticleService) Enabled() bool {
	return normalizeLLMCredential(global.GVA_CONFIG.VolcArk.ApiKey) != ""
}

func (s *VolcArkArticleService) apiKey() string {
	return normalizeLLMCredential(global.GVA_CONFIG.VolcArk.ApiKey)
}

func (s *VolcArkArticleService) chatURL() string {
	u := strings.TrimSpace(global.GVA_CONFIG.VolcArk.ChatURL)
	if u == "" {
		return volcArkChatURLDefault
	}
	return u
}

func (s *VolcArkArticleService) model() string {
	return strings.TrimSpace(global.GVA_CONFIG.VolcArk.Model)
}

func (s *VolcArkArticleService) httpTimeout() time.Duration {
	sec := global.GVA_CONFIG.VolcArk.Timeout
	if sec <= 0 {
		sec = 120
	}
	return time.Duration(sec) * time.Second
}

func (s *VolcArkArticleService) httpClient() *http.Client {
	return &http.Client{Timeout: s.httpTimeout()}
}

func (s *VolcArkArticleService) ensureConfig() error {
	if s.apiKey() == "" {
		return errors.New("请在 server/config.yaml 中配置 volc-ark.api-key（火山方舟 API Key，用作 Authorization: Bearer）")
	}
	if s.model() == "" {
		return errors.New("请配置 volc-ark.model。" + volcArkEndpointHint)
	}
	return nil
}

func (s *VolcArkArticleService) postChat(userContent string, temperature float64) (status int, body []byte, err error) {
	payload := map[string]interface{}{
		"model":       s.model(),
		"temperature": temperature,
		"messages": []map[string]string{
			{"role": "user", "content": userContent},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, s.chatURL(), bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey())
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

// GenerateArticle 调用豆包/方舟 chat.completions
func (s *VolcArkArticleService) GenerateArticle(req BaiduGenerateArticleReq) (string, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return "", errors.New("标题不能为空")
	}
	if err := articleLLMCheckRulesFileLen(req.RulesFileContent); err != nil {
		return "", err
	}
	if err := s.ensureConfig(); err != nil {
		return "", err
	}
	temp := req.Temperature
	if temp <= 0 {
		temp = 0.7
	}
	if temp > 1.0 {
		temp = 1.0
	}
	st, body, err := s.postChat(BuildArticleLLMPrompt(req), temp)
	if err != nil {
		return "", err
	}
	if st != http.StatusOK {
		global.GVA_LOG.Warn("火山方舟接口非 200", zap.Int("status", st), zap.String("body", string(body)))
		if text, perr := parseOpenAIChatCompletion(body, st, "火山方舟"); perr == nil && text != "" {
			return text, nil
		}
		return "", volcArkWrapHTTPError(st, body)
	}
	text, err := parseOpenAIChatCompletion(body, st, "火山方舟")
	if err != nil {
		return "", err
	}
	if text == "" {
		return "", errors.New("火山方舟返回正文为空，请调整提示词或稍后重试")
	}
	return text, nil
}

// Diagnose 最小请求探测连通与鉴权
func (s *VolcArkArticleService) Diagnose() ArticleLLMDiagnoseResult {
	out := ArticleLLMDiagnoseResult{
		Provider:         "volc-ark",
		AuthMode:         "ark-bearer",
		ModelEndpoint:    s.chatURL(),
		ModelServiceName: s.model(),
		TokenNote:        "使用火山方舟 API Key（Authorization: Bearer）",
	}
	key := s.apiKey()
	if len(key) >= 4 {
		out.ApiKeyTail = key[len(key)-4:]
	}
	if err := s.ensureConfig(); err != nil {
		out.ConfigDetail = err.Error()
		out.SuggestedAction = "在 server/config.yaml 中配置 volc-ark.api-key。"
		return out
	}
	out.ConfigOK = true
	out.TokenOK = true

	st, body, err := s.postChat("ping", 0.1)
	if err != nil {
		out.ChatErrorMsg = err.Error()
		out.SuggestedAction = "请求未到达火山或 TLS 失败：检查服务器出网、代理。"
		return out
	}
	out.ChatHTTPStatus = st
	if st != http.StatusOK {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.SuggestedAction = fmt.Sprintf("方舟 HTTP %d：核对 chat-url、API Key、以及 model 是否为接入点 ep- ID。", st)
		if st == http.StatusNotFound && bytes.Contains(body, []byte("InvalidEndpointOrModel")) {
			out.SuggestedAction = volcArkEndpointHint
		}
		if _, perr := parseOpenAIChatCompletion(body, st, "火山方舟"); perr != nil {
			out.ChatErrorMsg = perr.Error()
		}
		return out
	}
	_, perr := parseOpenAIChatCompletion(body, st, "火山方舟")
	if perr != nil {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.ChatErrorMsg = perr.Error()
		out.SuggestedAction = "响应解析失败或业务错误：查看 chatBodySnippet。"
		return out
	}
	out.ChatOK = true
	out.SuggestedAction = "火山方舟 Bearer 与 chat/completions 调用正常。"
	return out
}
