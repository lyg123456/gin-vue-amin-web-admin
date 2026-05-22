package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/flipped-aurora/gin-vue-admin/server/service/videoasync"
)

// VideoAsyncWorker 注册 Processor 并启动 goroutine + channel Worker
func VideoAsyncWorker() {
	if !global.GVA_CONFIG.VideoAsync.Enabled {
		return
	}
	videoasync.RegisterProcessor(func(shortVideoID uint) error {
		return content.ExecuteVideoGeneration(shortVideoID)
	})
	videoasync.Default().Start()
}
