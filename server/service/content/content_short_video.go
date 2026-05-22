package content

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"gorm.io/gorm"
)

type ShortVideoService struct{}

func suggestShortVideoSlug(title string) string {
	raw := strings.TrimSpace(title)
	if raw == "" {
		return fmt.Sprintf("sv-%d", time.Now().Unix())
	}
	s := strings.ToLower(raw)
	s = strings.NewReplacer(" ", "-").Replace(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) >= 3 {
		if len(out) > 120 {
			out = out[:120]
		}
		return out
	}
	return fmt.Sprintf("sv-%d", time.Now().Unix())
}

func (s *ShortVideoService) Create(v *contentModel.ContentShortVideo) error {
	v.Title = strings.TrimSpace(v.Title)
	if v.Title == "" {
		return errors.New("标题不能为空")
	}
	v.Slug = strings.TrimSpace(v.Slug)
	if v.Slug == "" {
		v.Slug = suggestShortVideoSlug(v.Title)
	}
	v.DurationSec = clampDuration(v.DurationSec)
	SyncShortVideoFrames(v)
	if v.Status == "" {
		v.Status = "draft"
	}
	return global.GVA_DB.Create(v).Error
}

func (s *ShortVideoService) Update(v *contentModel.ContentShortVideo) error {
	if v.ID == 0 {
		return errors.New("ID 不能为空")
	}
	v.Title = strings.TrimSpace(v.Title)
	if v.Title == "" {
		return errors.New("标题不能为空")
	}
	v.Slug = strings.TrimSpace(v.Slug)
	if v.Slug == "" {
		v.Slug = suggestShortVideoSlug(v.Title)
	}
	v.DurationSec = clampDuration(v.DurationSec)
	SyncShortVideoFrames(v)
	return global.GVA_DB.Save(v).Error
}

func (s *ShortVideoService) Delete(id uint) error {
	if id == 0 {
		return errors.New("ID 不能为空")
	}
	return global.GVA_DB.Delete(&contentModel.ContentShortVideo{}, id).Error
}

func (s *ShortVideoService) Find(id uint) (contentModel.ContentShortVideo, error) {
	var v contentModel.ContentShortVideo
	err := global.GVA_DB.Where("id = ?", id).First(&v).Error
	return v, err
}

func (s *ShortVideoService) GetList(page request.PageInfo, status string) (list []contentModel.ContentShortVideo, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.ContentShortVideo{}).Order("id desc")
	if strings.TrimSpace(status) != "" {
		db = db.Where("status = ?", status)
	}
	if strings.TrimSpace(page.Keyword) != "" {
		kw := "%" + strings.TrimSpace(page.Keyword) + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(page.Paginate()).Find(&list).Error
	return
}

func (s *ShortVideoService) Publish(id uint) error {
	if id == 0 {
		return errors.New("ID 不能为空")
	}
	now := time.Now()
	var v contentModel.ContentShortVideo
	if err := global.GVA_DB.Where("id = ?", id).First(&v).Error; err != nil {
		return err
	}
	if strings.TrimSpace(v.VideoURL) == "" {
		return errors.New("请先上传或生成成片视频后再发布")
	}
	return global.GVA_DB.Model(&contentModel.ContentShortVideo{}).Where("id = ?", id).Updates(map[string]any{
		"status":       "published",
		"published_at": &now,
	}).Error
}

func (s *ShortVideoService) GetPublishedList(page request.PageInfo) (list []contentModel.ContentShortVideo, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.ContentShortVideo{}).
		Where("status = ?", "published").
		Order("sort desc, published_at desc, id desc")
	if strings.TrimSpace(page.Keyword) != "" {
		kw := "%" + strings.TrimSpace(page.Keyword) + "%"
		db = db.Where("title LIKE ? OR description LIKE ?", kw, kw)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Select("id", "title", "slug", "description", "cover_image", "video_url", "duration_sec", "view_count", "published_at", "created_at", "updated_at").
		Scopes(page.Paginate()).Find(&list).Error
	return
}

func (s *ShortVideoService) GetPublishedBySlug(slug string) (contentModel.ContentShortVideo, error) {
	var v contentModel.ContentShortVideo
	err := global.GVA_DB.Where("slug = ? AND status = ?", slug, "published").First(&v).Error
	return v, err
}

func (s *ShortVideoService) IncViewBySlug(slug string) error {
	return global.GVA_DB.Model(&contentModel.ContentShortVideo{}).
		Where("slug = ? AND status = ?", slug, "published").
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

func (s *ShortVideoService) SetGenerating(id uint) error {
	return global.GVA_DB.Model(&contentModel.ContentShortVideo{}).Where("id = ?", id).Updates(map[string]any{
		"status":            "generating",
		"generation_error":  "",
	}).Error
}

func (s *ShortVideoService) SetGenerationResult(id uint, status, videoURL, taskID, genErr string) error {
	updates := map[string]any{
		"status": status,
	}
	if videoURL != "" {
		updates["video_url"] = videoURL
	}
	if taskID != "" {
		updates["generation_task_id"] = taskID
	}
	if genErr != "" {
		updates["generation_error"] = genErr
	} else {
		updates["generation_error"] = ""
	}
	return global.GVA_DB.Model(&contentModel.ContentShortVideo{}).Where("id = ?", id).Updates(updates).Error
}

func (s *ShortVideoService) RunVideoGeneration(id uint) error {
	if global.GVA_CONFIG.VideoAsync.Enabled {
		_, err := EnqueueVideoGeneration(id)
		return err
	}
	return ExecuteVideoGeneration(id)
}
