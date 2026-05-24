package content

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GetMediaCapabilities(c *gin.Context) {
	response.OkWithDetailed(officeMediaService.Capabilities(), "ok", c)
}

func (a *OfficeToolsApi) ProcessMedia(c *gin.Context) {
	action := c.PostForm("action")
	fh, _ := c.FormFile("file")
	data, filename, mime, err := officeMediaService.ProcessVideo(action, fh)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mime, data)
}

func (a *OfficeToolsApi) CompositeImages(c *gin.Context) {
	bg, _ := c.FormFile("background")
	fg, _ := c.FormFile("foreground")
	posX, _ := strconv.Atoi(c.DefaultPostForm("x", "0"))
	posY, _ := strconv.Atoi(c.DefaultPostForm("y", "0"))
	scale, _ := strconv.Atoi(c.DefaultPostForm("scale", "100"))
	data, filename, mime, err := officeMediaService.CompositeImages(bg, fg, posX, posY, scale)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mime, data)
}

func (a *OfficeToolsApi) ExtractImageBackground(c *gin.Context) {
	fh, _ := c.FormFile("file")
	data, filename, mime, err := officeMediaService.ExtractImageBackground(fh)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mime, data)
}

func (a *OfficeToolsApi) ProxyMediaDownload(c *gin.Context) {
	var req content.MediaDownloadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := content.StreamMediaDownload(req, c.Writer); err != nil {
		if !c.Writer.Written() {
			response.FailWithMessage(err.Error(), c)
		}
		return
	}
}
