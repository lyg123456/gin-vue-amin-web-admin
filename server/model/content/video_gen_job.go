package content

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// VideoGenJob 短视频成片异步任务（Redis 队列 + Worker 消费）
type VideoGenJob struct {
	global.GVA_MODEL

	ShortVideoID uint `json:"shortVideoId" gorm:"index;comment:短视频ID"`

	// queued / processing / succeeded / failed
	Status string `json:"status" gorm:"type:varchar(20);index;default:queued;comment:任务状态"`

	Provider       string `json:"provider" gorm:"type:varchar(20);comment:dashscope/volc"`
	ExternalTaskID string `json:"externalTaskId" gorm:"type:varchar(100);comment:第三方任务ID"`
	ErrorMsg       string `json:"errorMsg" gorm:"type:varchar(500);comment:失败原因"`
	Attempts       uint   `json:"attempts" gorm:"default:0;comment:已执行次数"`

	EnqueuedAt *time.Time `json:"enqueuedAt" gorm:"comment:入队时间"`
	StartedAt  *time.Time `json:"startedAt" gorm:"comment:开始处理时间"`
	FinishedAt *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
}

func (VideoGenJob) TableName() string {
	return "content_video_gen_jobs"
}
