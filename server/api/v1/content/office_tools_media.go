package content

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
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
	posX := officeMediaServiceParseInt(c.PostForm("x"), 0)
	posY := officeMediaServiceParseInt(c.PostForm("y"), 0)
	scale := officeMediaServiceParseInt(c.PostForm("scale"), 100)
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

func officeMediaServiceParseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func (a *OfficeToolsApi) CompressImage(c *gin.Context) {
	fh, _ := c.FormFile("file")
	quality := officeMediaServiceParseInt(c.PostForm("quality"), 80)
	maxW := officeMediaServiceParseInt(c.PostForm("maxWidth"), 1920)
	res, err := officeCompressService.CompressImage(fh, quality, maxW)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, res.Filename))
	c.Header("X-Original-Size", fmt.Sprintf("%d", res.OriginalSize))
	c.Header("X-New-Size", fmt.Sprintf("%d", res.NewSize))
	c.Data(http.StatusOK, res.Mime, res.Data)
}

func (a *OfficeToolsApi) CompressExcel(c *gin.Context) {
	fh, _ := c.FormFile("file")
	res, err := officeCompressService.CompressExcel(fh)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, res.Filename))
	c.Header("X-Original-Size", fmt.Sprintf("%d", res.OriginalSize))
	c.Header("X-New-Size", fmt.Sprintf("%d", res.NewSize))
	c.Data(http.StatusOK, res.Mime, res.Data)
}

func (a *OfficeToolsApi) GetTweetStyles(c *gin.Context) {
	response.OkWithDetailed(officeTweetService.ListStyles(), "ok", c)
}

func (a *OfficeToolsApi) RewriteTweet(c *gin.Context) {
	var req struct {
		Text  string `json:"text"`
		Style string `json:"style"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	out, err := officeTweetService.Rewrite(req.Text, req.Style)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"text": out}, "ok", c)
}
