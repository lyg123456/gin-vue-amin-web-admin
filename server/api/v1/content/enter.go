package content

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	ArticleApi
}

var (
	articleService = service.ServiceGroupApp.ContentServiceGroup.ArticleService
)

