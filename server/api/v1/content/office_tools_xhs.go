package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GetXhsOfficialCategories(c *gin.Context) {
	response.OkWithData(officeXhsCrawlService.ListCategories(), c)
}

func (a *OfficeToolsApi) VerifyXhsCookie(c *gin.Context) {
	var req struct {
		Cookie string `json:"cookie"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(officeXhsCrawlService.VerifyCookie(req.Cookie), c)
}

func (a *OfficeToolsApi) CrawlXhsIndustryVideos(c *gin.Context) {
	var req content.XhsIndustryCrawlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, note, err := officeXhsCrawlService.CrawlByCookie(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": data, "note": note}, "抓取完成", c)
}
