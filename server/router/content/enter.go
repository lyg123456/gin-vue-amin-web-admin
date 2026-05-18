package content

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	ArticleRouter
	ShortVideoRouter
}

var (
	articleApi           = api.ApiGroupApp.ContentApiGroup.ArticleApi
	articleCategoryApi   = api.ApiGroupApp.ContentApiGroup.ArticleCategoryApi
	portalContactLeadApi = api.ApiGroupApp.ContentApiGroup.PortalContactLeadApi
	shortVideoApi        = api.ApiGroupApp.ContentApiGroup.ShortVideoApi
)

