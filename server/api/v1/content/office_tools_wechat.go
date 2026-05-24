package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/gin-gonic/gin"
)

func (a *OfficeToolsApi) GetWechatOfficialCategories(c *gin.Context) {
	response.OkWithData(officeWechatCrawlService.ListCategories(), c)
}

func (a *OfficeToolsApi) VerifyWechatCookie(c *gin.Context) {
	var req struct {
		Cookie        string `json:"cookie"`
		WxChannelBase string `json:"wxChannelBase"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(officeWechatCrawlService.VerifyCookieWithOptions(req.Cookie, req.WxChannelBase), c)
}

func (a *OfficeToolsApi) CrawlWechatIndustryVideos(c *gin.Context) {
	var req content.WechatIndustryCrawlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, note, err := officeWechatCrawlService.CrawlByCookie(req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": data, "note": note}, "抓取完成", c)
}
