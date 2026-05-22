package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"go.uber.org/zap"
)

// InitOfficeTools 启动时检测 LibreOffice，并清理过期临时数据
func InitOfficeTools() {
	content.LogLibreOfficeStatus()
	content.LogFFmpegStatus()
	content.LogOfficeDataPolicy()
	if n, err := content.CleanOfficeToolsExpired(); err != nil {
		global.GVA_LOG.Warn("办公工具启动清理失败", zap.Error(err))
	} else if n > 0 {
		global.GVA_LOG.Info("办公工具启动清理", zap.Int("removed", n))
	}
}
