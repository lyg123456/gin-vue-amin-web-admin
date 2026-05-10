package content

import (
	"gorm.io/gorm"
)

// SiteStats 网站访问统计
type SiteStats struct {
	gorm.Model
	Total int64 `json:"total" gorm:"default:0"` // 总访问量
	Today int64 `json:"today" gorm:"default:0"` // 今日访问量
}