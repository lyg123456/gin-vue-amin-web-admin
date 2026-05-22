package content

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GetWatermarkCapabilities(c *gin.Context) {
	response.OkWithDetailed(officeWatermarkService.Capabilities(), "ok", c)
}

func (a *OfficeToolsApi) RemoveWatermark(c *gin.Context) {
	fh, _ := c.FormFile("file")
	preset := c.DefaultPostForm("preset", "douyin")
	method := c.DefaultPostForm("method", "blur")
	cx := parseFloatForm(c.PostForm("x"), -1)
	cy := parseFloatForm(c.PostForm("y"), -1)
	cw := parseFloatForm(c.PostForm("w"), -1)
	ch := parseFloatForm(c.PostForm("h"), -1)

	data, filename, mime, err := officeWatermarkService.RemoveWatermark(fh, preset, method, cx, cy, cw, ch)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mime, data)
}

func parseFloatForm(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
