package content

import (
	"fmt"
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GenerateWebStyleZip(c *gin.Context) {
	var req struct {
		URL      string `json:"url"`
		MaxPages int    `json:"maxPages"`
		MaxDepth int    `json:"maxDepth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, filename, err := officeWebStyleService.DownloadWebsiteZIP(req.URL, req.MaxPages, req.MaxDepth)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/zip", data)
}

func (a *OfficeToolsApi) CrawlWebProductsExcel(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, filename, err := officeWebCrawlService.CrawlProductsExcel(req.URL)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
