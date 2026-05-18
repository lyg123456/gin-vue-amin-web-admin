package content

import (
	"strings"

	adapter "github.com/casbin/gorm-adapter/v3"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// SyncContentInit 将“内容获客(文章)”相关的菜单/API/Casbin 以增量方式写入当前数据库。
// 适用场景：项目已完成 InitDB，后续新增了业务菜单或权限，需要同步到已有库。
func SyncContentInit(db *gorm.DB) error {
	if db == nil {
		return errors.New("db 不能为空")
	}

	// -------- 1) 菜单（父：content，子：contentArticle）--------
	var parent sysModel.SysBaseMenu
	if err := db.Where("name = ?", "content").First(&parent).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		parent = sysModel.SysBaseMenu{
			MenuLevel: 0,
			Hidden:    false,
			ParentId:  0,
			Path:      "content",
			Name:      "content",
			Component: "view/routerHolder.vue",
			Sort:      2,
			Meta:      sysModel.Meta{Title: "内容获客", Icon: "document"},
		}
		if err := db.Create(&parent).Error; err != nil {
			return errors.Wrap(err, "创建内容获客父级菜单失败")
		}
	}

	var articleMenu sysModel.SysBaseMenu
	if err := db.Where("name = ?", "contentArticle").First(&articleMenu).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		articleMenu = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      "article",
			Name:      "contentArticle",
			Component: "view/content/article/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "文章管理", Icon: "document"},
		}
		if err := db.Create(&articleMenu).Error; err != nil {
			return errors.Wrap(err, "创建文章管理子菜单失败")
		}
	}

	var leadMenu sysModel.SysBaseMenu
	if err := db.Where("name = ?", "contentPortalContactLead").First(&leadMenu).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		leadMenu = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      "contactLead",
			Name:      "contentPortalContactLead",
			Component: "view/content/contactLead/index.vue",
			Sort:      2,
			Meta:      sysModel.Meta{Title: "访客留资", Icon: "iphone"},
		}
		if err := db.Create(&leadMenu).Error; err != nil {
			return errors.Wrap(err, "创建访客留资子菜单失败")
		}
	}

	var aiMenu sysModel.SysBaseMenu
	if err := db.Where("name = ?", "contentAiArticle").First(&aiMenu).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		aiMenu = sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      "aiArticle",
			Name:      "contentAiArticle",
			Component: "view/content/aiArticle/index.vue",
			Sort:      3,
			Meta:      sysModel.Meta{Title: "AI写文章", Icon: "edit"},
		}
		if err := db.Create(&aiMenu).Error; err != nil {
			return errors.Wrap(err, "创建AI写文章子菜单失败")
		}
	}

	// 菜单-角色关联：确保 888/8881/9528 都能看到
	for _, aid := range []uint{888, 8881, 9528} {
		var auth sysModel.SysAuthority
		if err := db.Where("authority_id = ?", aid).First(&auth).Error; err != nil {
			continue
		}
		if err := db.Model(&auth).Association("SysBaseMenus").Append([]sysModel.SysBaseMenu{parent, articleMenu, leadMenu, aiMenu}); err != nil {
			return errors.Wrap(err, "为角色追加内容获客菜单失败")
		}
	}

	// -------- 2) SysApi（用于权限管理/同步 API）--------
	apis := []sysModel.SysApi{
		{ApiGroup: "内容获客", Method: "POST", Path: "/contentArticle/createArticle", Description: "创建文章"},
		{ApiGroup: "内容获客", Method: "PUT", Path: "/contentArticle/updateArticle", Description: "更新文章"},
		{ApiGroup: "内容获客", Method: "DELETE", Path: "/contentArticle/deleteArticle", Description: "删除文章"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentArticle/findArticle", Description: "获取单篇文章"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentArticle/getArticleList", Description: "获取文章列表"},
		{ApiGroup: "内容获客", Method: "POST", Path: "/contentArticle/publishArticle", Description: "发布文章"},
		{ApiGroup: "内容获客", Method: "POST", Path: "/contentArticle/generateArticleByBaidu", Description: "百度文心AI写文章"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentArticle/diagnoseBaiduWenxin", Description: "百度文心连接诊断"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentArticleCategory/getCategoryTree", Description: "文章分类树"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentArticleCategory/getCategoryList", Description: "文章分类列表"},
		{ApiGroup: "内容获客", Method: "POST", Path: "/contentArticleCategory/createCategory", Description: "创建文章分类"},
		{ApiGroup: "内容获客", Method: "PUT", Path: "/contentArticleCategory/updateCategory", Description: "更新文章分类"},
		{ApiGroup: "内容获客", Method: "DELETE", Path: "/contentArticleCategory/deleteCategory", Description: "删除文章分类"},
		{ApiGroup: "内容获客", Method: "GET", Path: "/contentPortalContactLead/getPortalContactLeadList", Description: "访客留资列表"},
		{ApiGroup: "系统初始化", Method: "POST", Path: "/contentInit/sync", Description: "同步内容获客初始化数据"},
	}
	for _, api := range apis {
		var cnt int64
		if err := db.Model(&sysModel.SysApi{}).Where("path = ? AND method = ?", api.Path, api.Method).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		if err := db.Create(&api).Error; err != nil {
			return errors.Wrap(err, "写入内容获客API失败")
		}
	}

	// -------- 3) Casbin（接口权限）--------
	paths := []struct {
		method string
		path   string
	}{
		{"POST", "/contentArticle/createArticle"},
		{"PUT", "/contentArticle/updateArticle"},
		{"DELETE", "/contentArticle/deleteArticle"},
		{"GET", "/contentArticle/findArticle"},
		{"GET", "/contentArticle/getArticleList"},
		{"POST", "/contentArticle/publishArticle"},
		{"POST", "/contentArticle/generateArticleByBaidu"},
		{"GET", "/contentArticle/diagnoseBaiduWenxin"},
		{"GET", "/contentArticleCategory/getCategoryTree"},
		{"GET", "/contentArticleCategory/getCategoryList"},
		{"POST", "/contentArticleCategory/createCategory"},
		{"PUT", "/contentArticleCategory/updateCategory"},
		{"DELETE", "/contentArticleCategory/deleteCategory"},
		{"GET", "/contentPortalContactLead/getPortalContactLeadList"},
		{"POST", "/contentInit/sync"},
	}
	for _, role := range []string{"888", "8881", "9528"} {
		for _, pth := range paths {
			rule := adapter.CasbinRule{Ptype: "p", V0: role, V1: pth.path, V2: pth.method}
			var cnt int64
			if err := db.Model(&adapter.CasbinRule{}).
				Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).
				Count(&cnt).Error; err != nil {
				return err
			}
			if cnt > 0 {
				continue
			}
			if err := db.Create(&rule).Error; err != nil {
				return errors.Wrap(err, "写入内容获客Casbin规则失败")
			}
		}
	}

	// -------- 3b) Casbin：凡已有「创建文章」权限的角色，自动补「AI 写文章」接口（兼容自定义角色 ID）--------
	var createArticleRoles []adapter.CasbinRule
	if err := db.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/contentArticle/createArticle", "POST").
		Find(&createArticleRoles).Error; err != nil {
		return errors.Wrap(err, "查询创建文章权限角色失败")
	}
	for _, row := range createArticleRoles {
		rid := strings.TrimSpace(row.V0)
		if rid == "" {
			continue
		}
		aiRule := adapter.CasbinRule{Ptype: "p", V0: rid, V1: "/contentArticle/generateArticleByBaidu", V2: "POST"}
		var cnt int64
		if err := db.Model(&adapter.CasbinRule{}).
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", aiRule.Ptype, aiRule.V0, aiRule.V1, aiRule.V2).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		if err := db.Create(&aiRule).Error; err != nil {
			return errors.Wrap(err, "补充AI写文章Casbin规则失败")
		}
	}

	// -------- 3c) Casbin：凡已有「AI 写文章」权限的角色，补「文心诊断」GET --------
	var genBaiduRoles []adapter.CasbinRule
	if err := db.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/contentArticle/generateArticleByBaidu", "POST").
		Find(&genBaiduRoles).Error; err != nil {
		return errors.Wrap(err, "查询AI写文章权限角色失败")
	}
	for _, row := range genBaiduRoles {
		rid := strings.TrimSpace(row.V0)
		if rid == "" {
			continue
		}
		diagRule := adapter.CasbinRule{Ptype: "p", V0: rid, V1: "/contentArticle/diagnoseBaiduWenxin", V2: "GET"}
		var cnt int64
		if err := db.Model(&adapter.CasbinRule{}).
			Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", diagRule.Ptype, diagRule.V0, diagRule.V1, diagRule.V2).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		if err := db.Create(&diagRule).Error; err != nil {
			return errors.Wrap(err, "为已有AI写文章权限角色补充文心诊断失败")
		}
	}

	// -------- 4) 默认文章分类（表为空时写入示例一级/二级）--------
	var catCnt int64
	if err := db.Model(&contentModel.ContentArticleCategory{}).Count(&catCnt).Error; err != nil {
		return err
	}
	if catCnt == 0 {
		root1 := contentModel.ContentArticleCategory{ParentID: 0, Name: "新闻动态", Sort: 1, Enabled: true}
		if err := db.Create(&root1).Error; err != nil {
			return errors.Wrap(err, "写入默认文章分类失败")
		}
		root2 := contentModel.ContentArticleCategory{ParentID: 0, Name: "文档教程", Sort: 2, Enabled: true}
		if err := db.Create(&root2).Error; err != nil {
			return errors.Wrap(err, "写入默认文章分类失败")
		}
		if err := db.Create(&contentModel.ContentArticleCategory{ParentID: root1.ID, Name: "公司动态", Sort: 1, Enabled: true}).Error; err != nil {
			return errors.Wrap(err, "写入默认文章子分类失败")
		}
		if err := db.Create(&contentModel.ContentArticleCategory{ParentID: root2.ID, Name: "使用指南", Sort: 1, Enabled: true}).Error; err != nil {
			return errors.Wrap(err, "写入默认文章子分类失败")
		}
	}

	return nil
}

