package content

import (
	"context"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"gorm.io/gorm"
)

type VideoGenJobService struct{}

func (s *VideoGenJobService) HasPending(shortVideoID uint) (bool, error) {
	var n int64
	err := global.GVA_DB.Model(&contentModel.VideoGenJob{}).
		Where("short_video_id = ? AND status IN ?", shortVideoID, []string{"queued", "processing"}).
		Count(&n).Error
	return n > 0, err
}

func (s *VideoGenJobService) CreateQueued(shortVideoID uint, provider string) (*contentModel.VideoGenJob, error) {
	now := time.Now()
	job := &contentModel.VideoGenJob{
		ShortVideoID: shortVideoID,
		Status:       "queued",
		Provider:     provider,
		Attempts:     0,
		EnqueuedAt:   &now,
	}
	return job, global.GVA_DB.Create(job).Error
}

func (s *VideoGenJobService) Find(id uint) (contentModel.VideoGenJob, error) {
	var j contentModel.VideoGenJob
	err := global.GVA_DB.Where("id = ?", id).First(&j).Error
	return j, err
}

func (s *VideoGenJobService) FindLatestByShortVideo(shortVideoID uint) (contentModel.VideoGenJob, error) {
	var j contentModel.VideoGenJob
	err := global.GVA_DB.Where("short_video_id = ?", shortVideoID).
		Order("id desc").First(&j).Error
	return j, err
}

func (s *VideoGenJobService) MarkProcessing(id uint) error {
	now := time.Now()
	return global.GVA_DB.Model(&contentModel.VideoGenJob{}).Where("id = ?", id).Updates(map[string]any{
		"status":      "processing",
		"started_at":  &now,
		"attempts":    gorm.Expr("attempts + 1"),
	}).Error
}

func (s *VideoGenJobService) MarkFinished(id uint, status, externalTaskID, errMsg string) error {
	now := time.Now()
	updates := map[string]any{
		"status":       status,
		"finished_at":  &now,
		"error_msg":    errMsg,
	}
	if externalTaskID != "" {
		updates["external_task_id"] = externalTaskID
	}
	return global.GVA_DB.Model(&contentModel.VideoGenJob{}).Where("id = ?", id).Updates(updates).Error
}

func (s *VideoGenJobService) GetList(page request.PageInfo, status string, shortVideoID uint) (list []contentModel.VideoGenJob, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.VideoGenJob{}).Order("id desc")
	if st := strings.TrimSpace(status); st != "" {
		db = db.Where("status = ?", st)
	}
	if shortVideoID > 0 {
		db = db.Where("short_video_id = ?", shortVideoID)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(page.Paginate()).Find(&list).Error
	return list, total, err
}

func (s *VideoGenJobService) QueueStats() (redisLen int64, err error) {
	if global.GVA_REDIS == nil {
		return 0, nil
	}
	ctx := context.Background()
	key := global.GVA_CONFIG.VideoAsync.QueueKey
	if key == "" {
		key = "gva:video:gen:queue"
	}
	return global.GVA_REDIS.LLen(ctx, key).Result()
}

var videoGenJobService VideoGenJobService
