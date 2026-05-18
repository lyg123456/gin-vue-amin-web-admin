package content

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const articleLLMRulesFileMaxRunes = 32000

// normalizeLLMCredential 去掉 YAML 里误加的空格、引号（百度 / 火山等共用）
func normalizeLLMCredential(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return strings.TrimSpace(s)
}

// BuildArticleLLMPrompt 组装「AI 写文章」用户提示词（百度 / 火山共用）
func BuildArticleLLMPrompt(req BaiduGenerateArticleReq) string {
	wc := req.WordCount
	if wc <= 0 {
		wc = 1000
	}
	if wc > 8000 {
		wc = 8000
	}
	doc := strings.TrimSpace(req.RulesFileContent)
	var b strings.Builder
	b.WriteString("你是专业的新媒体与 SEO 内容作者。请根据以下约束写一篇中文文章。\n\n")

	if doc != "" {
		b.WriteString("【最高优先级】以下为《用户规则文件》全文。你必须完整理解并严格执行其中的结构、语气、禁忌、段落要求等；")
		b.WriteString("若规则文件与后文「补充说明」冲突，以规则文件为准；规则文件未约定处，再采用后文默认约束。\n\n")
		b.WriteString("<<<BEGIN_RULES_FILE>>>\n")
		b.WriteString(doc)
		b.WriteString("\n<<<END_RULES_FILE>>>\n\n")
	}

	b.WriteString(fmt.Sprintf("【核心标题】%s\n", strings.TrimSpace(req.Title)))
	kw := strings.TrimSpace(req.Keywords)
	if kw != "" {
		b.WriteString(fmt.Sprintf("【关键词/主题侧重（可选补充）】%s\n", kw))
	}
	rules := strings.TrimSpace(req.Rules)
	if rules != "" {
		b.WriteString(fmt.Sprintf("【额外写作要求（可选补充）】%s\n", rules))
	}
	b.WriteString(fmt.Sprintf("【篇幅】约 %d 字（可适当浮动 ±15%%）。\n", wc))
	if doc == "" {
		b.WriteString("【格式与风格】干货、通俗，适合公众号；使用 Markdown：用 ## / ### 分级标题，列表用 - ，重点可加粗；不要输出代码块演示无关编程；结尾用一小段「总结」。\n")
	} else {
		b.WriteString("【格式】若规则文件已规定格式（如标题层级、是否 Markdown），从其规定；否则使用 Markdown：## / ### 分级标题，列表用 - ，结尾有简要总结。\n")
	}
	b.WriteString("【输出】只输出文章正文；不要输出「好的」「以下是」等套话；不要重复「核心标题」作为文章内的一级标题。\n")
	return b.String()
}

func llmBodySnippet(b []byte, max int) string {
	if max <= 0 {
		return ""
	}
	str := string(b)
	if len(str) <= max {
		return str
	}
	return str[:max] + "…"
}

// openAIChatCompletion 兼容千帆 V2 / 火山方舟等 chat.completions JSON
type openAIChatCompletion struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string          `json:"message"`
		Type    string          `json:"type"`
		Code    json.RawMessage `json:"code"`
	} `json:"error"`
}

func parseOpenAIChatCompletion(body []byte, httpStatus int, errPrefix string) (string, error) {
	var v openAIChatCompletion
	if err := json.Unmarshal(body, &v); err != nil {
		return "", fmt.Errorf("%s 解析响应失败: %w；HTTP=%d 片段=%s", errPrefix, err, httpStatus, llmBodySnippet(body, 280))
	}
	if v.Error != nil && strings.TrimSpace(v.Error.Message) != "" {
		code := strings.TrimSpace(string(v.Error.Code))
		if code != "" {
			return "", fmt.Errorf("%s: %s (code=%s)", errPrefix, v.Error.Message, code)
		}
		return "", fmt.Errorf("%s: %s", errPrefix, v.Error.Message)
	}
	if len(v.Choices) > 0 {
		out := strings.TrimSpace(v.Choices[0].Message.Content)
		if out != "" {
			return out, nil
		}
	}
	return "", fmt.Errorf("%s 返回正文为空（HTTP=%d）", errPrefix, httpStatus)
}

// ArticleLLMDiagnoseResult 一键诊断（百度 / 火山共用 JSON 形态）
type ArticleLLMDiagnoseResult struct {
	Provider         string `json:"provider"` // volc-ark | baidu-v2-bearer | baidu-oauth-access-token
	AuthMode         string `json:"authMode"`
	ConfigOK         bool   `json:"configOk"`
	ConfigDetail     string `json:"configDetail,omitempty"`
	ApiKeyTail       string `json:"apiKeyTail,omitempty"`
	ModelEndpoint    string `json:"modelEndpoint"`
	ModelServiceName string `json:"modelServiceName"`
	TokenNote        string `json:"tokenNote,omitempty"`
	TokenOK          bool   `json:"tokenOk"`
	TokenExpiresInSec int    `json:"tokenExpiresInSec,omitempty"`
	TokenError       string `json:"tokenError,omitempty"`
	ChatHTTPStatus   int    `json:"chatHttpStatus"`
	ChatOK           bool   `json:"chatOk"`
	ChatErrorCode    int    `json:"chatErrorCode"`
	ChatErrorMsg     string `json:"chatErrorMsg,omitempty"`
	ChatBodySnippet  string `json:"chatBodySnippet,omitempty"`
	SuggestedAction  string `json:"suggestedAction,omitempty"`
}

func articleLLMCheckRulesFileLen(doc string) error {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return nil
	}
	n := utf8.RuneCountInString(doc)
	if n > articleLLMRulesFileMaxRunes {
		return fmt.Errorf("规则文件内容过长（当前约 %d 字），请精简至不超过约 %d 字", n, articleLLMRulesFileMaxRunes)
	}
	return nil
}
