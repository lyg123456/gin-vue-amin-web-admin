package content

import (
	"errors"
	"strings"

	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
)

// NormalizeFrameURLs 合并显式首尾帧与旧版 sourceImages（逗号分隔，前两位为首/尾）
func NormalizeFrameURLs(first, last, sourceImages string) (string, string) {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first != "" && last != "" {
		return first, last
	}
	for _, u := range parseSourceImageURLs(sourceImages) {
		if first == "" {
			first = u
			continue
		}
		if last == "" {
			last = u
			break
		}
	}
	return first, last
}

func JoinFrameURLs(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first == "" && last == "" {
		return ""
	}
	if last == "" {
		return first
	}
	if first == "" {
		return last
	}
	return first + "," + last
}

func SyncShortVideoFrames(v *contentModel.ContentShortVideo) {
	if v == nil {
		return
	}
	first, last := NormalizeFrameURLs(v.FirstFrameURL, v.LastFrameURL, v.SourceImages)
	v.FirstFrameURL = first
	v.LastFrameURL = last
	v.SourceImages = JoinFrameURLs(first, last)
}

func ValidateI2VFrameURLs(first, last string) error {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first == "" {
		return errors.New("请上传首帧图（first_frame，公网可访问 URL）")
	}
	if last == "" {
		return errors.New("请上传尾帧图（last_frame，公网可访问 URL）")
	}
	return nil
}
