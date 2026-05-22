package content

import (
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	contentModel "github.com/flipped-aurora/gin-vue-admin/server/model/content"
	"gorm.io/gorm"
)

type PortalVisitorService struct{}

func todayDateString() string {
	return time.Now().Format("2006-01-02")
}

func normalizeVisitorIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "unknown"
	}
	if len(ip) > 64 {
		ip = ip[:64]
	}
	return ip
}

// RecordVisit 记录一次门户访问（按 IP + 自然日累加）
func (s *PortalVisitorService) RecordVisit(ip string) error {
	ip = normalizeVisitorIP(ip)
	date := todayDateString()

	var row contentModel.PortalVisitorDaily
	err := global.GVA_DB.Where("client_ip = ? AND visit_date = ?", ip, date).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = contentModel.PortalVisitorDaily{
			ClientIP:   ip,
			VisitDate:  date,
			VisitCount: 1,
		}
		return global.GVA_DB.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return global.GVA_DB.Model(&row).UpdateColumn("visit_count", gorm.Expr("visit_count + ?", 1)).Error
}

type PortalVisitorSummary struct {
	VisitDate    string `json:"visitDate"`
	TotalPV      int64  `json:"totalPv"`      // 历史总访问次数（全表 visit_count 之和）
	TotalUV      int64  `json:"totalUv"`      // 历史独立 IP 数
	TodayPV      int64  `json:"todayPv"`
	TodayUV      int64  `json:"todayUv"`
	FilterDatePV int64  `json:"filterDatePv"`
	FilterDateUV int64  `json:"filterDateUv"`
}

func (s *PortalVisitorService) GetSummary(visitDate string) (PortalVisitorSummary, error) {
	today := todayDateString()
	if visitDate == "" {
		visitDate = today
	}
	out := PortalVisitorSummary{VisitDate: visitDate}

	var totalPV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Select("COALESCE(SUM(visit_count),0)").Scan(&totalPV).Error; err != nil {
		return out, err
	}
	var totalUV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Distinct("client_ip").Count(&totalUV).Error; err != nil {
		return out, err
	}
	out.TotalPV = totalPV
	out.TotalUV = totalUV

	var todayPV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Where("visit_date = ?", today).
		Select("COALESCE(SUM(visit_count),0)").Scan(&todayPV).Error; err != nil {
		return out, err
	}
	var todayUV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Where("visit_date = ?", today).Count(&todayUV).Error; err != nil {
		return out, err
	}
	out.TodayPV = todayPV
	out.TodayUV = todayUV

	var filterPV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Where("visit_date = ?", visitDate).
		Select("COALESCE(SUM(visit_count),0)").Scan(&filterPV).Error; err != nil {
		return out, err
	}
	var filterUV int64
	if err := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{}).
		Where("visit_date = ?", visitDate).Count(&filterUV).Error; err != nil {
		return out, err
	}
	out.FilterDatePV = filterPV
	out.FilterDateUV = filterUV
	return out, nil
}

func (s *PortalVisitorService) GetList(info request.PageInfo, visitDate, keyword string) (list []contentModel.PortalVisitorDaily, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.PortalVisitorDaily{})
	if visitDate = strings.TrimSpace(visitDate); visitDate != "" {
		db = db.Where("visit_date = ?", visitDate)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		db = db.Where("client_ip LIKE ?", "%"+kw+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Order("visit_date desc, visit_count desc, id desc").
		Scopes(info.Paginate()).
		Find(&list).Error
	return list, total, err
}
