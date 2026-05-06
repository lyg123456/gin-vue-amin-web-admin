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

type ArticleService struct{}

func (s *ArticleService) CreateArticle(article *contentModel.ContentArticle) error {
	article.Slug = strings.TrimSpace(article.Slug)
	article.Title = strings.TrimSpace(article.Title)
	if article.Slug == "" || article.Title == "" {
		return errors.New("title/slug 不能为空")
	}
	return global.GVA_DB.Create(article).Error
}

func (s *ArticleService) UpdateArticle(article *contentModel.ContentArticle) error {
	if article.ID == 0 {
		return errors.New("ID 不能为空")
	}
	article.Slug = strings.TrimSpace(article.Slug)
	article.Title = strings.TrimSpace(article.Title)
	if article.Slug == "" || article.Title == "" {
		return errors.New("title/slug 不能为空")
	}
	return global.GVA_DB.Save(article).Error
}

func (s *ArticleService) DeleteArticle(id uint, authorID uint) error {
	if id == 0 {
		return errors.New("ID 不能为空")
	}
	db := global.GVA_DB.Model(&contentModel.ContentArticle{}).Where("id = ?", id)
	if authorID > 0 {
		db = db.Where("author_id = ?", authorID)
	}
	return db.Delete(&contentModel.ContentArticle{}).Error
}

func (s *ArticleService) FindArticle(id uint, authorID uint) (contentModel.ContentArticle, error) {
	var a contentModel.ContentArticle
	db := global.GVA_DB.Where("id = ?", id)
	if authorID > 0 {
		db = db.Where("author_id = ?", authorID)
	}
	err := db.First(&a).Error
	return a, err
}

func (s *ArticleService) GetArticleList(authorID uint, page request.PageInfo, status string) (list []contentModel.ContentArticle, total int64, err error) {
	db := global.GVA_DB.Model(&contentModel.ContentArticle{}).Order("id desc")
	if authorID > 0 {
		db = db.Where("author_id = ?", authorID)
	}
	if strings.TrimSpace(status) != "" {
		db = db.Where("status = ?", status)
	}
	if strings.TrimSpace(page.Keyword) != "" {
		kw := "%" + strings.TrimSpace(page.Keyword) + "%"
		db = db.Where("title LIKE ? OR slug LIKE ?", kw, kw)
	}
	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Scopes(page.Paginate()).Find(&list).Error
	return
}

func (s *ArticleService) PublishArticle(id uint, authorID uint) error {
	if id == 0 {
		return errors.New("ID 不能为空")
	}
	now := time.Now()
	db := global.GVA_DB.Model(&contentModel.ContentArticle{}).Where("id = ?", id)
	if authorID > 0 {
		db = db.Where("author_id = ?", authorID)
	}
	return db.Updates(map[string]any{
		"status":       "published",
		"published_at": &now,
	}).Error
}

func (s *ArticleService) GetPublishedList(page request.PageInfo) (list []contentModel.ContentArticle, total int64, err error) {
	q := global.GVA_DB.Model(&contentModel.ContentArticle{}).Where("status = ?", "published")
	if strings.TrimSpace(page.Keyword) != "" {
		kw := "%" + strings.TrimSpace(page.Keyword) + "%"
		q = q.Where("title LIKE ? OR summary LIKE ?", kw, kw)
	}
	err = q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = q.Session(&gorm.Session{}).
		Omit("content").
		Order("published_at DESC, id DESC").
		Scopes(page.Paginate()).
		Find(&list).Error
	return
}

func (s *ArticleService) GetPublishedBySlug(slug string) (contentModel.ContentArticle, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return contentModel.ContentArticle{}, errors.New("slug 不能为空")
	}
	var a contentModel.ContentArticle
	err := global.GVA_DB.Where("slug = ? AND status = ?", slug, "published").First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contentModel.ContentArticle{}, err
	}
	return a, err
}

func (s *ArticleService) IncViewBySlug(slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	return global.GVA_DB.Model(&contentModel.ContentArticle{}).
		Where("slug = ? AND status = ?", slug, "published").
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

