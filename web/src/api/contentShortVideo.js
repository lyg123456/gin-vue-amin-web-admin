import service from '@/utils/request'

export const getShortVideoList = (params) => {
  return service({
    url: '/contentShortVideo/getShortVideoList',
    method: 'get',
    params
  })
}

export const findShortVideo = (params) => {
  return service({
    url: '/contentShortVideo/findShortVideo',
    method: 'get',
    params
  })
}

export const createShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/createShortVideo',
    method: 'post',
    data
  })
}

export const updateShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/updateShortVideo',
    method: 'put',
    data
  })
}

export const deleteShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/deleteShortVideo',
    method: 'delete',
    data
  })
}

export const publishShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/publishShortVideo',
    method: 'post',
    data
  })
}

export const generateShortVideoScript = (data) => {
  return service({
    url: '/contentShortVideo/generateShortVideoScript',
    method: 'post',
    data,
    timeout: 600000
  })
}

export const createShortVideoWithAI = (data) => {
  return service({
    url: '/contentShortVideo/createShortVideoWithAI',
    method: 'post',
    data,
    timeout: 600000
  })
}

export const generateShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/generateShortVideo',
    method: 'post',
    data,
    timeout: 600000
  })
}

export const regenerateShortVideo = (data) => {
  return service({
    url: '/contentShortVideo/regenerateShortVideo',
    method: 'post',
    data,
    timeout: 600000
  })
}

export const getPublishedShortVideos = (params) => {
  return service({
    url: '/public/shortVideos',
    method: 'get',
    params,
    donNotShowLoading: true
  })
}

export const getPublishedShortVideoBySlug = (slug) => {
  return service({
    url: `/public/shortVideo/${encodeURIComponent(slug)}`,
    method: 'get',
    donNotShowLoading: true
  })
}
