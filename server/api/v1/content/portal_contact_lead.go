package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PortalContactLeadApi struct{}

type portalContactLeadSubmitReq struct {
	Phone  string `json:"phone"`
	Remark string `json:"remark"`
}

// Submit 门户公开提交留资
func (a *PortalContactLeadApi) Submit(c *gin.Context) {
	var req portalContactLeadSubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	lead := contentModel.PortalContactLead{
		Phone:    req.Phone,
		Remark:   req.Remark,
		ClientIP: c.ClientIP(),
	}
	if err := portalContactLeadService.Submit(&lead); err != nil {
		global.GVA_LOG.Warn("门户留资提交失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("提交成功，我们会尽快与您联系", c)
}

// GetList 后台分页列表
func (a *PortalContactLeadApi) GetList(c *gin.Context) {
	var pageInfo request.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := portalContactLeadService.GetList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取留资列表失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{
		"list":     list,
		"total":    total,
		"page":     pageInfo.Page,
		"pageSize": pageInfo.PageSize,
	}, "获取成功", c)
}
