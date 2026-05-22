package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	contentRouter "github.com/flipped-aurora/gin-vue-admin/server/router/content"
	"github.com/gin-gonic/gin"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	holder(publicGroup, privateGroup)

	// 内容发布（文章 + SEO）
	new(contentRouter.RouterGroup).InitContentArticleRouter(privateGroup, publicGroup)
	new(contentRouter.RouterGroup).InitPortalVisitorRouter(privateGroup)
	new(contentRouter.RouterGroup).InitContentShortVideoRouter(privateGroup, publicGroup)

}
