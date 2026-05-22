import service from '@/utils/request'

export const getVideoGenJobList = (params) => {
  return service({
    url: '/contentVideoGenJob/getVideoGenJobList',
    method: 'get',
    params
  })
}
