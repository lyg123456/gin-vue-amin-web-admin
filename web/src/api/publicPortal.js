import service from '@/utils/request'

/** 公开：门户首页轮播图（无登录） */
export const getPublicHomeCarousel = () => {
  return service({
    url: '/public/homeCarousel',
    method: 'get',
    donNotShowLoading: true
  })
}
