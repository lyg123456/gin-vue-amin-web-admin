package content

type ServiceGroup struct {
	ArticleService
	ArticleCategoryService
	PortalCarouselService
	PortalContactLeadService
	BaiduWenxinArticleService
	VolcArkArticleService
	ShortVideoService
	VolcArkShortVideoService
	VolcArkVideoGenerateService
	DashScopeVideoService
}

var (
	volcArkArticleService       VolcArkArticleService
	volcArkVideoGenerateService VolcArkVideoGenerateService
	shortVideoService           ShortVideoService
	volcArkShortVideoService    VolcArkShortVideoService
	dashScopeVideoService       DashScopeVideoService
)

// articleCategoryService 供同包 ArticleService 调用（避免循环依赖 ServiceGroup）
var articleCategoryService ArticleCategoryService

