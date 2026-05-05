package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/content"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		&content.ContentArticle{},
	)
	if err != nil {
		return err
	}
	return nil
}
