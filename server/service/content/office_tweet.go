package content

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type OfficeTweetService struct{}

var tweetStyles = map[string]string{
	"default":  "简洁有力、适合微博/推特",
	"xhs":      "小红书风格，口语化、带适量 emoji",
	"wechat":   "公众号风格，段落清晰、有感染力",
	"formal":   "正式克制、商务表达",
	"humorous": "轻松幽默、有记忆点",
}

// Rewrite 推文 AI 改写（优先火山方舟，其次百度文心）
func (s *OfficeTweetService) Rewrite(text, style string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("请输入推文内容")
	}
	styleHint := tweetStyles[style]
	if styleHint == "" {
		styleHint = tweetStyles["default"]
	}
	prompt := fmt.Sprintf(`你是社交媒体文案专家。请对下列推文进行改写润色。
要求：%s；保留核心信息；直接输出改写后的正文，不要解释、不要加标题。
原文：
%s`, styleHint, text)

	if volcArkArticleService.Enabled() {
		if err := volcArkArticleService.ensureConfig(); err == nil {
			st, body, err := volcArkArticleService.postChat(prompt, 0.75)
			if err == nil && st == http.StatusOK {
				if out, perr := parseOpenAIChatCompletion(body, st, "火山方舟"); perr == nil && strings.TrimSpace(out) != "" {
					return strings.TrimSpace(out), nil
				}
			}
		}
	}
	baidu := &BaiduWenxinArticleService{}
	if err := baidu.ensureConfig(); err == nil {
		if baidu.usesV2() {
			st, body, err := baidu.postV2Chat(prompt, 0.75)
			if err == nil && st == http.StatusOK {
				if out, perr := parseOpenAIChatCompletion(body, st, "百度文心"); perr == nil && strings.TrimSpace(out) != "" {
					return strings.TrimSpace(out), nil
				}
			}
		}
	}
	return "", errors.New("AI 改写不可用：请在 config.yaml 配置 volc-ark.api-key 或 baidu-wenxin")
}

// ListStyles 返回可用风格
func (s *OfficeTweetService) ListStyles() map[string]string {
	return tweetStyles
}
