package content

// PortalCarouselService 门户首页轮播（数据可由后续改为读库或配置中心）
type PortalCarouselService struct{}

// HomeCarouselSlide 单帧数据，与前端字段一致
type HomeCarouselSlide struct {
	Image    string `json:"image"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	LinkText string `json:"linkText"`
	LinkUrl  string `json:"linkUrl"`
}

// GetSlides 返回首页轮播列表（无鉴权接口使用；后续可替换为数据库查询）
func (s *PortalCarouselService) GetSlides() []HomeCarouselSlide {
	return []HomeCarouselSlide{
		{
			Image:    "https://images.unsplash.com/photo-1497215728101-856f4a433174?w=1600&q=80",
			Title:    "专业 专注 建站系统首选方案",
			Subtitle: "基于前后端分离的站点内容发布与浏览能力，可扩展会员与营销模块。",
			LinkText: "浏览文章",
			LinkUrl:  "/",
		},
		{
			Image:    "https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?w=1600&q=80",
			Title:    "内容驱动 · 稳定可靠",
			Subtitle: "文章、SEO 与公开访问接口开箱即用，可按业务继续定制。",
			LinkText: "会员中心",
			LinkUrl:  "/member",
		},
		{
			Image:    "https://images.unsplash.com/photo-1522071820081-009f0129c71c?w=1600&q=80",
			Title:    "开放生态 · 易于扩展",
			Subtitle: "Gin-Vue-Admin 业务模块与插件机制，便于团队二次开发。",
			LinkText: "了解更多",
			LinkUrl:  "https://github.com/flipped-aurora/gin-vue-admin",
		},
	}
}
