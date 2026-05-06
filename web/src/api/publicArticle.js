import service from '@/utils/request'

/** 公开：已发布文章分页列表（无正文） */
export const getPublicArticleList = (params) => {
  return service({
    url: '/public/contentArticles',
    method: 'get',
    params,
    donNotShowLoading: true
  })
}

/** 公开：按 slug 获取已发布文章全文 */
export const getPublicArticleBySlug = (slug) => {
  return service({
    url: `/public/article/${encodeURIComponent(slug)}`,
    method: 'get',
    donNotShowLoading: true
  })
}
