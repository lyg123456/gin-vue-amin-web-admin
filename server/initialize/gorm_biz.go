package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/content"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		&content.ContentArticle{},
		&content.ContentArticleCategory{},
		&content.PortalContactLead{},
		&content.ContentShortVideo{},
	)
	if err != nil {
		return err
	}
	ensureBizTablesUTF8MB4(db)
	return nil
}
