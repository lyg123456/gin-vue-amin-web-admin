package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GetDouyinOfficialCategories(c *gin.Context) {
	response.OkWithData(officeDouyinCrawlService.ListCategories(), c)
}

func (a *OfficeToolsApi) VerifyDouyinCookie(c *gin.Context) {
	var req struct {
		Cookie string `json:"cookie"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(officeDouyinCrawlService.VerifyCookie(req.Cookie), c)
}

func (a *OfficeToolsApi) CrawlDouyinIndustryVideos(c *gin.Context) {
	var req content.DouyinIndustryCrawlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, note, err := officeDouyinCrawlService.CrawlByCookie(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": data, "note": note}, "抓取完成", c)
}
