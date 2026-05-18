package content

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ContentShortVideo 短视频获客：AI 脚本 + 成片入库
type ContentShortVideo struct {
	global.GVA_MODEL

	AuthorID uint `json:"authorId" form:"authorId" gorm:"index;comment:创建人ID"`

	Title       string `json:"title" form:"title" gorm:"type:varchar(200);charset:utf8mb4;collation:utf8mb4_unicode_ci;index;comment:标题"`
	Slug        string `json:"slug" form:"slug" gorm:"type:varchar(200);charset:utf8mb4;collation:utf8mb4_unicode_ci;uniqueIndex;comment:门户访问slug"`
	Description string `json:"description" form:"description" gorm:"type:varchar(1000);charset:utf8mb4;collation:utf8mb4_unicode_ci;comment:视频描述/场景说明"`
	PromptText  string `json:"promptText" form:"promptText" gorm:"type:text;charset:utf8mb4;collation:utf8mb4_unicode_ci;comment:用户输入的创意文字"`
	Script      string `json:"script" form:"script" gorm:"type:longtext;charset:utf8mb4;collation:utf8mb4_unicode_ci;comment:AI/人工分镜文案"`

	DurationSec uint `json:"durationSec" form:"durationSec" gorm:"default:60;comment:目标时长秒(30-120)"`

	CoverImage    string `json:"coverImage" form:"coverImage" gorm:"type:varchar(500);comment:封面图URL"`
	FirstFrameURL string `json:"firstFrameUrl" form:"firstFrameUrl" gorm:"type:varchar(500);comment:DashScope首帧图URL"`
	LastFrameURL  string `json:"lastFrameUrl" form:"lastFrameUrl" gorm:"type:varchar(500);comment:DashScope尾帧图URL"`
	SourceImages  string `json:"sourceImages" form:"sourceImages" gorm:"type:text;comment:首尾帧逗号拼接(兼容)"`
	VideoURL     string `json:"videoUrl" form:"videoUrl" gorm:"type:varchar(500);comment:成片视频URL"`

	// draft / generating / ready / failed / published / archived
	Status           string `json:"status" form:"status" gorm:"type:varchar(20);index;default:draft;comment:状态"`
	GenerationError  string `json:"generationError" form:"generationError" gorm:"type:varchar(500);comment:生成失败原因"`
	AiProvider       string `json:"aiProvider" form:"aiProvider" gorm:"type:varchar(20);default:volc;comment:脚本AI来源"`
	GenerationTaskID string `json:"generationTaskId" form:"generationTaskId" gorm:"type:varchar(100);comment:火山视频任务ID(预留)"`

	PublishedAt *time.Time `json:"publishedAt" form:"publishedAt" gorm:"index;comment:发布时间"`
	ViewCount   uint       `json:"viewCount" form:"viewCount" gorm:"comment:播放量"`
	Sort        int        `json:"sort" form:"sort" gorm:"default:0;comment:门户排序"`
}

func (ContentShortVideo) TableName() string {
	return "content_short_videos"
}
