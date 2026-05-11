package content

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ContentArticle 用于“信息发布 + SEO获客”的文章实体
// - 后台：企业/用户登录后发布文章
// - 前台：通过 slug 公开访问（用于 SEO 收录）
type ContentArticle struct {
	global.GVA_MODEL

	AuthorID uint `json:"authorId" form:"authorId" gorm:"index;comment:作者用户ID"`

	// CategoryID 关联 content_article_categories；0 表示未分类
	CategoryID uint                      `json:"categoryId" form:"categoryId" gorm:"index;default:0;comment:分类ID"`
	Category   *ContentArticleCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`

	Title   string `json:"title" form:"title" gorm:"type:varchar(200);index;comment:标题"`
	Slug    string `json:"slug" form:"slug" gorm:"type:varchar(200);uniqueIndex;comment:SEO友好链接slug"`
	Summary string `json:"summary" form:"summary" gorm:"type:varchar(500);comment:摘要"`

	Content     string `json:"content" form:"content" gorm:"type:longtext;comment:正文内容"`
	ContentType string `json:"contentType" form:"contentType" gorm:"type:varchar(20);default:markdown;comment:内容类型 markdown/html"`

	// cover_image：多图时存为逗号分隔 URL，最多 6 张（与前端约定）
	CoverImage string `json:"coverImage" form:"coverImage" gorm:"type:text;comment:封面图(多图逗号分隔)"`

	SEOTitle       string `json:"seoTitle" form:"seoTitle" gorm:"type:varchar(200);comment:SEO标题"`
	SEOKeywords    string `json:"seoKeywords" form:"seoKeywords" gorm:"type:varchar(500);comment:SEO关键词(逗号分隔)"`
	SEODescription string `json:"seoDescription" form:"seoDescription" gorm:"type:varchar(500);comment:SEO描述"`

	Status      string     `json:"status" form:"status" gorm:"type:varchar(20);index;default:draft;comment:状态 draft/published/archived"`
	PublishedAt *time.Time `json:"publishedAt" form:"publishedAt" gorm:"index;comment:发布时间"`

	ViewCount uint `json:"viewCount" form:"viewCount" gorm:"comment:浏览量"`
	LeadCount uint `json:"leadCount" form:"leadCount" gorm:"comment:线索量(后续迭代)"`
}

func (ContentArticle) TableName() string {
	return "content_articles"
}

