import service from '@/utils/request'

/** 门户公开：提交留资（电话 + 备注），无需登录 */
export const submitPortalContactLead = (data) => {
  return service({
    url: '/public/portalContactLead',
    method: 'post',
    data,
    donNotShowLoading: true
  })
}
