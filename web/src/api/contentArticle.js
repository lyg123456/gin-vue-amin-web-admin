import service from '@/utils/request'

export const createContentArticle = (data) => {
  return service({
    url: '/contentArticle/createArticle',
    method: 'post',
    data
  })
}

export const updateContentArticle = (data) => {
  return service({
    url: '/contentArticle/updateArticle',
    method: 'put',
    data
  })
}

export const deleteContentArticle = (data) => {
  return service({
    url: '/contentArticle/deleteArticle',
    method: 'delete',
    data
  })
}

export const findContentArticle = (params) => {
  return service({
    url: '/contentArticle/findArticle',
    method: 'get',
    params
  })
}

export const getContentArticleList = (params) => {
  return service({
    url: '/contentArticle/getArticleList',
    method: 'get',
    params
  })
}

export const publishContentArticle = (data) => {
  return service({
    url: '/contentArticle/publishArticle',
    method: 'post',
    data
  })
}

// public: SEO 公开访问（不需要 token，仍然走同一个 baseURL）
export const getPublishedArticleBySlug = (slug) => {
  return service({
    url: `/public/article/${encodeURIComponent(slug)}`,
    method: 'get'
  })
}

