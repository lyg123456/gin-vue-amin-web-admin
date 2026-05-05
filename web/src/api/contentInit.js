import service from '@/utils/request'

export const syncContentInit = () => {
  return service({
    url: '/contentInit/sync',
    method: 'post'
  })
}

