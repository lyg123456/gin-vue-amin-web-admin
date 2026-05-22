package content

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type VideoGenJobApi struct{}

func (a *VideoGenJobApi) GetVideoGenJobList(c *gin.Context) {
	var page request.PageInfo
	if err := c.ShouldBindQuery(&page); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status := c.Query("status")
	var shortVideoID uint
	if id := c.Query("shortVideoId"); id != "" {
		n, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			response.FailWithMessage("shortVideoId 无效", c)
			return
		}
		shortVideoID = uint(n)
	}
	list, total, err := videoGenJobService.GetList(page, status, shortVideoID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	redisLen, _ := videoGenJobService.QueueStats()
	response.OkWithDetailed(gin.H{
		"list":           list,
		"total":          total,
		"page":           page.Page,
		"pageSize":       page.PageSize,
		"redisQueueLen":  redisLen,
		"asyncEnabled":   global.GVA_CONFIG.VideoAsync.Enabled,
	}, "获取成功", c)
}
