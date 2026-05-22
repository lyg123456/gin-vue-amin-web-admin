package content

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// SyncShortVideoInit 增量同步「短视频获客」菜单 / API / Casbin
func SyncShortVideoInit(db *gorm.DB) error {
	if db == nil {
		return errors.New("db 不能为空")
	}

	var parent sysModel.SysBaseMenu
	if err := db.Where("name = ?", "shortVideo").First(&parent).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		parent = sysModel.SysBaseMenu{
			MenuLevel: 0,
			Hidden:    false,
			ParentId:  0,
			Path:      "shortVideo",
			Name:      "shortVideo",
			Component: "view/routerHolder.vue",
			Sort:      25,
			Meta:      sysModel.Meta{Title: "短视频获客", Icon: "video-camera"},
		}
		if err := db.Create(&parent).Error; err != nil {
			return errors.Wrap(err, "创建短视频获客父级菜单失败")
		}
	}

	menus := []struct {
		name, path, component, title, icon string
		sort                               int
	}{
		{"shortVideoAi", "aiGenerate", "view/shortVideo/aiGenerate/index.vue", "AI生成短视频", "magic-stick", 1},
		{"shortVideoList", "list", "view/shortVideo/list/index.vue", "短视频列表", "list", 2},
		{"shortVideoGenJobs", "jobs", "view/shortVideo/jobs/index.vue", "异步成片任务", "timer", 3},
	}
	for _, m := range menus {
		var cnt int64
		if err := db.Model(&sysModel.SysBaseMenu{}).Where("name = ?", m.name).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		item := sysModel.SysBaseMenu{
			MenuLevel: 1,
			Hidden:    false,
			ParentId:  parent.ID,
			Path:      m.path,
			Name:      m.name,
			Component: m.component,
			Sort:      m.sort,
			Meta:      sysModel.Meta{Title: m.title, Icon: m.icon},
		}
		if err := db.Create(&item).Error; err != nil {
			return errors.Wrap(err, "创建短视频子菜单失败")
		}
	}

	apis := []sysModel.SysApi{
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/createShortVideo", Description: "创建短视频"},
		{ApiGroup: "短视频获客", Method: "PUT", Path: "/contentShortVideo/updateShortVideo", Description: "更新短视频"},
		{ApiGroup: "短视频获客", Method: "DELETE", Path: "/contentShortVideo/deleteShortVideo", Description: "删除短视频"},
		{ApiGroup: "短视频获客", Method: "GET", Path: "/contentShortVideo/findShortVideo", Description: "短视频详情"},
		{ApiGroup: "短视频获客", Method: "GET", Path: "/contentShortVideo/getShortVideoList", Description: "短视频列表"},
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/publishShortVideo", Description: "发布短视频"},
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/generateShortVideoScript", Description: "AI生成短视频脚本"},
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/createShortVideoWithAI", Description: "AI创建短视频入库"},
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/generateShortVideo", Description: "生成短视频成片"},
		{ApiGroup: "短视频获客", Method: "POST", Path: "/contentShortVideo/regenerateShortVideo", Description: "重新生成短视频"},
		{ApiGroup: "短视频获客", Method: "GET", Path: "/contentVideoGenJob/getVideoGenJobList", Description: "异步成片任务列表"},
		{ApiGroup: "短视频获客", Method: "GET", Path: "/public/shortVideos", Description: "公开短视频列表"},
		{ApiGroup: "短视频获客", Method: "GET", Path: "/public/shortVideo/:slug", Description: "公开短视频详情"},
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
			return errors.Wrap(err, "写入短视频API失败")
		}
	}

	paths := []struct {
		method string
		path   string
	}{
		{"POST", "/contentShortVideo/createShortVideo"},
		{"PUT", "/contentShortVideo/updateShortVideo"},
		{"DELETE", "/contentShortVideo/deleteShortVideo"},
		{"GET", "/contentShortVideo/findShortVideo"},
		{"GET", "/contentShortVideo/getShortVideoList"},
		{"POST", "/contentShortVideo/publishShortVideo"},
		{"POST", "/contentShortVideo/generateShortVideoScript"},
		{"POST", "/contentShortVideo/createShortVideoWithAI"},
		{"POST", "/contentShortVideo/generateShortVideo"},
		{"POST", "/contentShortVideo/regenerateShortVideo"},
		{"GET", "/contentVideoGenJob/getVideoGenJobList"},
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
				return errors.Wrap(err, "写入短视频Casbin失败")
			}
		}
	}

	var aiMenu, listMenu, jobsMenu sysModel.SysBaseMenu
	_ = db.Where("name = ?", "shortVideoAi").First(&aiMenu).Error
	_ = db.Where("name = ?", "shortVideoList").First(&listMenu).Error
	_ = db.Where("name = ?", "shortVideoGenJobs").First(&jobsMenu).Error
	for _, aid := range []uint{888, 8881, 9528} {
		var auth sysModel.SysAuthority
		if err := db.Where("authority_id = ?", aid).First(&auth).Error; err != nil {
			continue
		}
		toAppend := []sysModel.SysBaseMenu{}
		if aiMenu.ID > 0 {
			toAppend = append(toAppend, aiMenu)
		}
		if listMenu.ID > 0 {
			toAppend = append(toAppend, listMenu)
		}
		if jobsMenu.ID > 0 {
			toAppend = append(toAppend, jobsMenu)
		}
		if len(toAppend) > 0 {
			_ = db.Model(&auth).Association("SysBaseMenus").Append(toAppend)
		}
	}

	return nil
}
