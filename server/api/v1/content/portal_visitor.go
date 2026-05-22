package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type PortalVisitorApi struct{}

func (a *PortalVisitorApi) GetPortalVisitorList(c *gin.Context) {
	var page request.PageInfo
	if err := c.ShouldBindQuery(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	visitDate := c.Query("visitDate")
	keyword := c.Query("keyword")
	list, total, err := portalVisitorService.GetList(page, visitDate, keyword)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	summary, _ := portalVisitorService.GetSummary(visitDate)
	response.OkWithDetailed(gin.H{
		"list":     list,
		"total":    total,
		"page":     page.Page,
		"pageSize": page.PageSize,
		"summary":  summary,
	}, "获取成功", c)
}

func (a *PortalVisitorApi) GetPortalVisitorSummary(c *gin.Context) {
	visitDate := c.Query("visitDate")
	summary, err := portalVisitorService.GetSummary(visitDate)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(summary, "获取成功", c)
}
