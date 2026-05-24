package content

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) CompressImage(c *gin.Context) {
	fh, _ := c.FormFile("file")
	quality, _ := strconv.Atoi(c.DefaultPostForm("quality", "80"))
	maxWidth, _ := strconv.Atoi(c.DefaultPostForm("maxWidth", "1920"))
	res, err := officeCompressService.CompressImage(fh, quality, maxWidth)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, res.Filename))
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
	c.Data(http.StatusOK, res.Mime, res.Data)
}
