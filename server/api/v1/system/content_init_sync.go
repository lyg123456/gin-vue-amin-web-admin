package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ContentInitApi struct{}

// SyncContentInit
// @Tags     ContentInit
// @Summary  同步内容获客初始化数据（菜单/API/Casbin）
// @Security ApiKeyAuth
// @Produce  application/json
// @Success  200  {object}  response.Response{msg=string}  "同步成功"
// @Router   /contentInit/sync [post]
func (a *ContentInitApi) SyncContentInit(c *gin.Context) {
	if err := contentInitService.Sync(); err != nil {
		global.GVA_LOG.Error("同步内容获客初始化数据失败", zap.Error(err))
		response.FailWithMessage("同步失败", c)
		return
	}
	response.OkWithMessage("同步成功", c)
}

