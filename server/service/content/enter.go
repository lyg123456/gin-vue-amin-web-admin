package content

type ServiceGroup struct {
	ArticleService
	ArticleCategoryService
	PortalCarouselService
}

// articleCategoryService 供同包 ArticleService 调用（避免循环依赖 ServiceGroup）
var articleCategoryService ArticleCategoryService

