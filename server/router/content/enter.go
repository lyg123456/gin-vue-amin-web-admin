package content

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	ArticleRouter
	PortalVisitorRouter
	ShortVideoRouter
}

var (
	articleApi           = api.ApiGroupApp.ContentApiGroup.ArticleApi
	articleCategoryApi   = api.ApiGroupApp.ContentApiGroup.ArticleCategoryApi
	portalContactLeadApi = api.ApiGroupApp.ContentApiGroup.PortalContactLeadApi
	portalVisitorApi     = api.ApiGroupApp.ContentApiGroup.PortalVisitorApi
	shortVideoApi        = api.ApiGroupApp.ContentApiGroup.ShortVideoApi
	videoGenJobApi       = api.ApiGroupApp.ContentApiGroup.VideoGenJobApi
	officeToolsApi       = api.ApiGroupApp.ContentApiGroup.OfficeToolsApi
)

