package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// GetLibreOfficeStatus 检测 LibreOffice 是否可用（便于排查）
func (a *OfficeToolsApi) GetLibreOfficeStatus(c *gin.Context) {
	caps := officeFileConvertService.Capabilities()
	response.OkWithDetailed(gin.H{
		"libreOffice":      caps.LibreOffice,
		"libreOfficePath":  caps.LibreOfficePath,
		"goOfficeFallback": caps.GoOfficeFallback,
		"goFallbackExts":   caps.GoFallbackExts,
		"hint":             caps.LibreOfficeHint,
		"installScript":    "scripts/install-libreoffice.ps1",
		"configKey":        "office-tools.libreoffice-path",
	}, "ok", c)
}
