package initialize

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ensureBizTablesUTF8MB4 修复 MySQL utf8(3字节) 表无法存 emoji 等问题（Error 1366）
func ensureBizTablesUTF8MB4(db *gorm.DB) {
	if global.GVA_CONFIG.System.DbType != "mysql" {
		return
	}
	tables := []string{
		"content_short_videos",
		"content_articles",
		"content_article_categories",
		"portal_contact_leads",
		"portal_visitor_daily",
		"content_video_gen_jobs",
	}
	for _, table := range tables {
		sql := fmt.Sprintf(
			"ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
			table,
		)
		if err := db.Exec(sql).Error; err != nil {
			global.GVA_LOG.Warn("业务表 utf8mb4 转换跳过（表可能尚未创建）",
				zap.String("table", table),
				zap.Error(err),
			)
			continue
		}
		global.GVA_LOG.Info("业务表已转换为 utf8mb4", zap.String("table", table))
	}
}
