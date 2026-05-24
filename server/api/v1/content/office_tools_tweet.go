package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

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
