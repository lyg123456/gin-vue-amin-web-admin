package content

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	ArticleApi
	ArticleCategoryApi
	PortalContactLeadApi
	PortalVisitorApi
	ShortVideoApi
	VideoGenJobApi
	OfficeToolsApi
}

var (
	articleService             = service.ServiceGroupApp.ContentServiceGroup.ArticleService
	articleCategoryService     = service.ServiceGroupApp.ContentServiceGroup.ArticleCategoryService
	portalCarouselService      = service.ServiceGroupApp.ContentServiceGroup.PortalCarouselService
	portalContactLeadService   = service.ServiceGroupApp.ContentServiceGroup.PortalContactLeadService
	portalVisitorService       = service.ServiceGroupApp.ContentServiceGroup.PortalVisitorService
	baiduWenxinArticleService = service.ServiceGroupApp.ContentServiceGroup.BaiduWenxinArticleService
	volcArkArticleService      = service.ServiceGroupApp.ContentServiceGroup.VolcArkArticleService
	shortVideoService          = service.ServiceGroupApp.ContentServiceGroup.ShortVideoService
	videoGenJobService         = service.ServiceGroupApp.ContentServiceGroup.VideoGenJobService
	volcArkShortVideoService   = service.ServiceGroupApp.ContentServiceGroup.VolcArkShortVideoService
	officeTempEmailService       = service.ServiceGroupApp.ContentServiceGroup.OfficeTempEmailService
	officeFileConvertService     = service.ServiceGroupApp.ContentServiceGroup.OfficeFileConvertService
	officeMediaService           = service.ServiceGroupApp.ContentServiceGroup.OfficeMediaService
	officeCompressService        = service.ServiceGroupApp.ContentServiceGroup.OfficeCompressService
	officeTweetService           = service.ServiceGroupApp.ContentServiceGroup.OfficeTweetService
	officeWebStyleService        = service.ServiceGroupApp.ContentServiceGroup.OfficeWebStyleService
	officeWebCrawlService        = service.ServiceGroupApp.ContentServiceGroup.OfficeWebCrawlService
	officeSpeedService           = service.ServiceGroupApp.ContentServiceGroup.OfficeSpeedService
	officeWatermarkService       = service.ServiceGroupApp.ContentServiceGroup.OfficeWatermarkService
	officeDouyinCrawlService     = service.ServiceGroupApp.ContentServiceGroup.OfficeDouyinCrawlService
	officeWechatCrawlService     = service.ServiceGroupApp.ContentServiceGroup.OfficeWechatCrawlService
	officeXhsCrawlService      = service.ServiceGroupApp.ContentServiceGroup.OfficeXhsCrawlService
)

