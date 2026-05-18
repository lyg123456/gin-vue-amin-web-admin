package content

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	ArticleApi
	ArticleCategoryApi
	PortalContactLeadApi
	ShortVideoApi
}

var (
	articleService             = service.ServiceGroupApp.ContentServiceGroup.ArticleService
	articleCategoryService     = service.ServiceGroupApp.ContentServiceGroup.ArticleCategoryService
	portalCarouselService      = service.ServiceGroupApp.ContentServiceGroup.PortalCarouselService
	portalContactLeadService   = service.ServiceGroupApp.ContentServiceGroup.PortalContactLeadService
	baiduWenxinArticleService = service.ServiceGroupApp.ContentServiceGroup.BaiduWenxinArticleService
	volcArkArticleService      = service.ServiceGroupApp.ContentServiceGroup.VolcArkArticleService
	shortVideoService          = service.ServiceGroupApp.ContentServiceGroup.ShortVideoService
	volcArkShortVideoService   = service.ServiceGroupApp.ContentServiceGroup.VolcArkShortVideoService
)

