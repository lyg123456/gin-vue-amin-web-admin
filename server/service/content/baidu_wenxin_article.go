package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	baiduOAuthTokenURL = "https://aip.baidubce.com/oauth/2.0/token"
	baiduV2ChatURLDefault = "https://qianfan.baidubce.com/v2/chat/completions"
	baiduV2ModelDefault   = "ernie-speed-128k"
)

// validateBaiduOAuthClientID 千帆 OAuth 的 client_id 必须是「应用」里的 API Key，不能用百度云 IAM 的 bce-v3/…
func validateBaiduOAuthClientID(clientID string) error {
	if clientID == "" {
		return nil
	}
	lower := strings.ToLower(clientID)
	if strings.Contains(lower, "bce-v3/") || strings.HasPrefix(lower, "bce-v3") {
		return errors.New("「api-key」填成了百度云 IAM 密钥（含 bce-v3/）。请打开百度千帆控制台 → 应用接入 → 点进你的应用 → 复制「API Key」填到 api-key，「Secret Key」填到 secret-key（二者成对，用于 OAuth client_credentials）")
	}
	return nil
}

// BaiduWenxinArticleService 调用百度千帆文心生成文章正文（优先 V2 Bearer，否则 OAuth + aip）
type BaiduWenxinArticleService struct{}

type BaiduGenerateArticleReq struct {
	Title            string  `json:"title"`
	Keywords         string  `json:"keywords"`
	Rules            string  `json:"rules"`
	RulesFileContent string  `json:"rulesFileContent"`
	WordCount        int     `json:"wordCount"`
	Temperature      float64 `json:"temperature"`
}

type baiduOAuthResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type baiduChatResp struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Result    string `json:"result"`
}

var (
	baiduTokenMu           sync.Mutex
	baiduCachedToken       string
	baiduTokenExpireAt     time.Time
	baiduTokenSingleFlight singleflight.Group
)

func (s *BaiduWenxinArticleService) usesV2() bool {
	return normalizeLLMCredential(global.GVA_CONFIG.BaiduWenxin.V2APIKey) != ""
}

func (s *BaiduWenxinArticleService) v2APIKey() string {
	return normalizeLLMCredential(global.GVA_CONFIG.BaiduWenxin.V2APIKey)
}

func (s *BaiduWenxinArticleService) v2ChatURL() string {
	u := strings.TrimSpace(global.GVA_CONFIG.BaiduWenxin.V2ChatURL)
	if u == "" {
		return baiduV2ChatURLDefault
	}
	return u
}

func (s *BaiduWenxinArticleService) v2Model() string {
	m := strings.TrimSpace(global.GVA_CONFIG.BaiduWenxin.V2Model)
	if m == "" {
		return baiduV2ModelDefault
	}
	return m
}

func (s *BaiduWenxinArticleService) ensureConfig() error {
	w := global.GVA_CONFIG.BaiduWenxin
	if s.usesV2() {
		if s.v2APIKey() == "" {
			return errors.New("请在 server/config.yaml 中配置 baidu-wenxin.v2-api-key（千帆 V2 OpenAPI 的 API Key，用作 Authorization: Bearer）")
		}
		return nil
	}
	ak := normalizeLLMCredential(w.ApiKey)
	sk := normalizeLLMCredential(w.SecretKey)
	if ak == "" || sk == "" {
		return errors.New("未配置 baidu-wenxin.v2-api-key 时，请配置 api-key 与 secret-key（千帆应用 OAuth），或改用 v2-api-key 走 V2 接口")
	}
	if err := validateBaiduOAuthClientID(ak); err != nil {
		return err
	}
	return nil
}

func (s *BaiduWenxinArticleService) httpTimeout() time.Duration {
	sec := global.GVA_CONFIG.BaiduWenxin.Timeout
	if sec <= 0 {
		sec = 120
	}
	return time.Duration(sec) * time.Second
}

func (s *BaiduWenxinArticleService) httpClient() *http.Client {
	return &http.Client{Timeout: s.httpTimeout()}
}

func (s *BaiduWenxinArticleService) modelEndpoint() string {
	ep := strings.TrimSpace(global.GVA_CONFIG.BaiduWenxin.ModelEndpoint)
	if ep == "" {
		return "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/ernie-speed-128k"
	}
	return ep
}

func baiduModelServiceName(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func (s *BaiduWenxinArticleService) fetchAccessToken() (string, time.Time, int, error) {
	if err := s.ensureConfig(); err != nil {
		return "", time.Time{}, 0, err
	}
	if s.usesV2() {
		return "", time.Time{}, 0, errors.New("内部错误：V2 模式不应请求 OAuth token")
	}
	w := global.GVA_CONFIG.BaiduWenxin
	ak := normalizeLLMCredential(w.ApiKey)
	sk := normalizeLLMCredential(w.SecretKey)
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", ak)
	form.Set("client_secret", sk)
	httpReq, err := http.NewRequest(http.MethodPost, baiduOAuthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", time.Time{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		return "", time.Time{}, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, 0, err
	}
	var o baiduOAuthResp
	if err := json.Unmarshal(body, &o); err != nil {
		return "", time.Time{}, 0, fmt.Errorf("解析 token 响应失败: %w", err)
	}
	if o.Error != "" {
		desc := strings.TrimSpace(o.ErrorDesc)
		msg := fmt.Sprintf("获取 token 失败: %s", o.Error)
		if desc != "" {
			msg += " " + desc
		}
		if strings.Contains(o.Error, "invalid_client") || strings.Contains(strings.ToLower(desc), "unknown client") {
			msg += "。按百度文档：unknown client id 表示「作为 client_id 的 API Key 不被 OAuth 识别」。请使用「AI开放能力 / 文心一言 → 控制台里创建的「应用」详情页」中的 API Key + Secret Key（access_token 鉴权那一套）；"
			msg += "不要填「安全认证 → API Key」里 bce-v3/ 形态的统一密钥，也不要填与 API Key 不成对、或来自其他应用的 Secret。"
		}
		return "", time.Time{}, 0, errors.New(msg)
	}
	if o.AccessToken == "" {
		return "", time.Time{}, 0, errors.New("获取 token 失败: 响应中无 access_token")
	}
	expiresIn := o.ExpiresIn
	exp := time.Now().Add(time.Duration(o.ExpiresIn) * time.Second)
	if o.ExpiresIn <= 0 {
		exp = time.Now().Add(50 * time.Minute)
	}
	return o.AccessToken, exp, expiresIn, nil
}

func (s *BaiduWenxinArticleService) getAccessToken() (string, error) {
	if s.usesV2() {
		return "", errors.New("内部错误：V2 模式不应调用 getAccessToken")
	}
	baiduTokenMu.Lock()
	if baiduCachedToken != "" && time.Now().Before(baiduTokenExpireAt.Add(-2*time.Minute)) {
		t := baiduCachedToken
		baiduTokenMu.Unlock()
		return t, nil
	}
	baiduTokenMu.Unlock()

	v, err, _ := baiduTokenSingleFlight.Do("baidu_access_token", func() (interface{}, error) {
		token, exp, _, err := s.fetchAccessToken()
		if err != nil {
			return "", err
		}
		baiduTokenMu.Lock()
		baiduCachedToken = token
		baiduTokenExpireAt = exp
		baiduTokenMu.Unlock()
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (s *BaiduWenxinArticleService) postV2Chat(userContent string, temperature float64) (status int, body []byte, err error) {
	payload := map[string]interface{}{
		"model":       s.v2Model(),
		"temperature": temperature,
		"messages": []map[string]string{
			{"role": "user", "content": userContent},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, s.v2ChatURL(), bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.v2APIKey())
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

func (s *BaiduWenxinArticleService) Diagnose() ArticleLLMDiagnoseResult {
	if s.usesV2() {
		return s.diagnoseV2()
	}
	return s.diagnoseOAuth()
}

func (s *BaiduWenxinArticleService) diagnoseV2() ArticleLLMDiagnoseResult {
	out := ArticleLLMDiagnoseResult{
		Provider:         "baidu-v2",
		AuthMode:         "v2-bearer",
		ModelEndpoint:    s.v2ChatURL(),
		ModelServiceName: s.v2Model(),
		TokenNote:        "V2 使用 Authorization: Bearer，无 OAuth access_token",
	}
	key := s.v2APIKey()
	if len(key) >= 4 {
		out.ApiKeyTail = key[len(key)-4:]
	}
	if err := s.ensureConfig(); err != nil {
		out.ConfigDetail = err.Error()
		out.SuggestedAction = "在 server/config.yaml 中配置 baidu-wenxin.v2-api-key（千帆控制台 V2 API Key）。"
		return out
	}
	out.ConfigOK = true
	out.TokenOK = true

	st, body, err := s.postV2Chat("ping", 0.1)
	if err != nil {
		out.ChatErrorMsg = err.Error()
		out.SuggestedAction = "请求未到达百度或 TLS 失败：检查服务器出网、代理。"
		return out
	}
	out.ChatHTTPStatus = st
	if st != http.StatusOK {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.SuggestedAction = fmt.Sprintf("V2 接口 HTTP %d：核对 v2-chat-url、Bearer 是否有效、账户是否开通该模型。", st)
		if _, perr := parseOpenAIChatCompletion(body, st, "千帆 V2"); perr != nil {
			out.ChatErrorMsg = perr.Error()
		}
		return out
	}
	text, perr := parseOpenAIChatCompletion(body, st, "千帆 V2")
	if perr != nil {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.ChatErrorMsg = perr.Error()
		out.SuggestedAction = "响应解析失败或业务错误：查看 chatBodySnippet / error 字段。"
		return out
	}
	_ = text
	out.ChatOK = true
	out.SuggestedAction = "V2 Bearer 与 chat/completions 调用正常。"
	return out
}

func (s *BaiduWenxinArticleService) diagnoseOAuth() ArticleLLMDiagnoseResult {
	out := ArticleLLMDiagnoseResult{
		Provider:         "baidu-oauth",
		AuthMode:         "oauth-access-token",
		ModelEndpoint:    s.modelEndpoint(),
		ModelServiceName: baiduModelServiceName(s.modelEndpoint()),
	}
	w := global.GVA_CONFIG.BaiduWenxin
	ak := normalizeLLMCredential(w.ApiKey)
	if len(ak) >= 4 {
		out.ApiKeyTail = ak[len(ak)-4:]
	}
	if err := s.ensureConfig(); err != nil {
		out.ConfigDetail = err.Error()
		out.SuggestedAction = "修正 server/config.yaml：配置 v2-api-key，或配置 api-key + secret-key（OAuth）。"
		return out
	}
	out.ConfigOK = true

	token, _, expSec, err := s.fetchAccessToken()
	if err != nil {
		out.TokenError = err.Error()
		out.SuggestedAction = "OAuth 失败：确认 API Key + Secret Key 来自同一千帆应用，且非 IAM（bce-v3）密钥。"
		return out
	}
	out.TokenOK = true
	out.TokenExpiresInSec = expSec

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"temperature": 0.1,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		out.ChatErrorMsg = err.Error()
		out.SuggestedAction = "本地序列化请求失败。"
		return out
	}
	postURL := out.ModelEndpoint + "?access_token=" + url.QueryEscape(token)
	httpReq, err := http.NewRequest(http.MethodPost, postURL, bytes.NewReader(raw))
	if err != nil {
		out.ChatErrorMsg = err.Error()
		out.SuggestedAction = "构造 HTTP 请求失败。"
		return out
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		out.ChatErrorMsg = err.Error()
		out.SuggestedAction = "请求未到达百度：检查服务器出网、DNS、防火墙、代理。"
		return out
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		out.ChatHTTPStatus = resp.StatusCode
		out.ChatErrorMsg = err.Error()
		return out
	}
	out.ChatHTTPStatus = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.SuggestedAction = fmt.Sprintf("对话接口 HTTP %d：核对 model-endpoint 是否为千帆「chat」类完整 URL。", resp.StatusCode)
		return out
	}
	var chat baiduChatResp
	if err := json.Unmarshal(body, &chat); err != nil {
		out.ChatBodySnippet = llmBodySnippet(body, 500)
		out.ChatErrorMsg = fmt.Sprintf("解析 JSON 失败: %v", err)
		out.SuggestedAction = "返回体不是预期对话 JSON：可能 model-endpoint 指到了非文心对话接口。"
		return out
	}
	out.ChatErrorCode = chat.ErrorCode
	out.ChatErrorMsg = strings.TrimSpace(chat.ErrorMsg)
	if chat.ErrorCode == 0 {
		out.ChatOK = true
		out.SuggestedAction = "OAuth 与当前模型对话均正常。若长文生成失败，可尝试减小规则文件或字数。"
		return out
	}
	out.ChatOK = false
	switch chat.ErrorCode {
	case 6:
		out.SuggestedAction = "错误 6：当前 endpoint 末尾服务名为「" + out.ModelServiceName + "」，该应用下可能未开通此在线推理服务，或与控制台已开通模型名不一致。请到千帆 → 应用 → 查看可调用的 chat 服务名，并把 config 中 model-endpoint 最后一段改成一致。"
	case 4, 17:
		out.SuggestedAction = "常与 QPS/并发或配额相关，请查看控制台用量与限流说明。"
	default:
		out.SuggestedAction = "请对照百度千帆/文心错误码文档，结合 error_msg 与账户计费状态排查。"
	}
	return out
}

// GenerateArticle 生成 Markdown 正文（V2：Bearer + chat/completions；旧版：OAuth + aip chat URL）
func (s *BaiduWenxinArticleService) GenerateArticle(req BaiduGenerateArticleReq) (string, error) {
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
	if s.usesV2() {
		return s.generateArticleV2(req, temp)
	}
	return s.generateArticleOAuth(req, temp)
}

func (s *BaiduWenxinArticleService) generateArticleV2(req BaiduGenerateArticleReq, temp float64) (string, error) {
	st, body, err := s.postV2Chat(BuildArticleLLMPrompt(req), temp)
	if err != nil {
		return "", err
	}
	if st != http.StatusOK {
		global.GVA_LOG.Warn("千帆 V2 接口非 200", zap.Int("status", st), zap.String("body", string(body)))
		if text, perr := parseOpenAIChatCompletion(body, st, "千帆 V2"); perr == nil && text != "" {
			return text, nil
		}
		return "", fmt.Errorf("千帆 V2 接口错误: HTTP %d %s", st, llmBodySnippet(body, 400))
	}
	text, err := parseOpenAIChatCompletion(body, st, "千帆 V2")
	if err != nil {
		return "", err
	}
	if text == "" {
		return "", errors.New("千帆 V2 返回正文为空，请调整提示词或稍后重试")
	}
	return text, nil
}

func (s *BaiduWenxinArticleService) generateArticleOAuth(req BaiduGenerateArticleReq, temp float64) (string, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return "", err
	}
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": BuildArticleLLMPrompt(req)},
		},
		"temperature": temp,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	ep := s.modelEndpoint()
	postURL := ep + "?access_token=" + url.QueryEscape(token)
	httpReq, err := http.NewRequest(http.MethodPost, postURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	resp, err := s.httpClient().Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		global.GVA_LOG.Warn("文心接口非 200", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("文心接口错误: HTTP %d", resp.StatusCode)
	}
	var chat baiduChatResp
	if err := json.Unmarshal(body, &chat); err != nil {
		return "", fmt.Errorf("解析文心响应失败: %w", err)
	}
	if chat.ErrorCode != 0 {
		msg := fmt.Sprintf("文心错误 %d: %s", chat.ErrorCode, chat.ErrorMsg)
		if chat.ErrorCode == 6 {
			msg += "。常见原因：① 千帆控制台该应用未开通/无额度调用当前「模型路径」对应的在线服务（如 ernie-speed-128k）；② 账号欠费或套餐不含该模型；③ `model-endpoint` 与控制台已开通的模型不一致。请到 https://console.bce.baidu.com/qianfan/ais/console/applicationConsole/application 打开应用 → 查看可调用的模型列表，将 `baidu-wenxin.model-endpoint` 改为与已开通模型一致的 chat 地址（路径最后一段为模型服务名）。"
		}
		return "", errors.New(msg)
	}
	out := strings.TrimSpace(chat.Result)
	if out == "" {
		return "", errors.New("文心返回正文为空，请调整提示词或稍后重试")
	}
	return out, nil
}
