package content

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	ArticleApi
	ArticleCategoryApi
}

var (
	articleService          = service.ServiceGroupApp.ContentServiceGroup.ArticleService
	articleCategoryService  = service.ServiceGroupApp.ContentServiceGroup.ArticleCategoryService
	portalCarouselService   = service.ServiceGroupApp.ContentServiceGroup.PortalCarouselService
)

