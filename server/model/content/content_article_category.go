package content

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// ContentArticleCategory 文章分类（最多二级：一级 parent_id=0，二级 parent_id 指向一级）
type ContentArticleCategory struct {
	global.GVA_MODEL

	ParentID uint   `json:"parentId" form:"parentId" gorm:"index;default:0;comment:父级ID，0为一级"`
	Name     string `json:"name" form:"name" gorm:"type:varchar(100);index;comment:分类名称"`
	Sort     int    `json:"sort" form:"sort" gorm:"default:0;comment:排序(小在前)"`
	Enabled  bool   `json:"enabled" form:"enabled" gorm:"default:true;comment:是否启用"`
}

func (ContentArticleCategory) TableName() string {
	return "content_article_categories"
}
