package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	contentSource "github.com/flipped-aurora/gin-vue-admin/server/source/content"
)

type ContentInitService struct{}

func (s *ContentInitService) Sync() error {
	if err := contentSource.SyncContentInit(global.GVA_DB); err != nil {
		return err
	}
	return contentSource.SyncShortVideoInit(global.GVA_DB)
}

