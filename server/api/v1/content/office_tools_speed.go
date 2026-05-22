package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) SpeedPing(c *gin.Context) {
	response.OkWithDetailed(officeSpeedService.Ping(), "ok", c)
}

func (a *OfficeToolsApi) SpeedInfo(c *gin.Context) {
	response.OkWithDetailed(officeSpeedService.ClientInfo(c), "ok", c)
}

func (a *OfficeToolsApi) SpeedDownload(c *gin.Context) {
	mb := officeSpeedService.ParseSizeMB(c, 1)
	officeSpeedService.WriteDownload(c, mb)
}

func (a *OfficeToolsApi) SpeedUpload(c *gin.Context) {
	res, err := officeSpeedService.UploadEcho(c)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "ok", c)
}
