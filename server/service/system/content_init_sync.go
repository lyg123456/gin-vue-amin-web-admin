package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	contentSource "github.com/flipped-aurora/gin-vue-admin/server/source/content"
)

type ContentInitService struct{}

func (s *ContentInitService) Sync() error {
	return contentSource.SyncContentInit(global.GVA_DB)
}

