package content

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// GetHomeCarousel 门户首页轮播图数据（公开接口，无需登录）
func (a *ArticleApi) GetHomeCarousel(c *gin.Context) {
	list := portalCarouselService.GetSlides()
	response.OkWithDetailed(gin.H{"list": list}, "获取成功", c)
}
