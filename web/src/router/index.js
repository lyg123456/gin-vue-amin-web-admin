import { createRouter, createWebHashHistory } from 'vue-router'

/**
 * 与部署路径一致：根路径 / 与用户端 /web/ → 门户；管理端 /admin/
 * 例：http://host/web/#/ 、http://host/admin/#/login
 */
export function getHistoryBase() {
  if (typeof window === 'undefined') return '/web/'
  const p = window.location.pathname || ''
  if (p === '/admin' || p.startsWith('/admin/')) return '/admin/'
  if (p === '/web' || p.startsWith('/web/')) return '/web/'
  // 根路径 / 等未带前缀的访问 → 用户端门户（与 Nginx 将 / 重定向到 /web/ 一致）
  return '/web/'
}

function buildRoutes() {
  const base = getHistoryBase()
  if (base === '/web/') {
    return [
      {
        path: '/',
        name: 'WebSite',
        meta: { title: '站点', client: true },
        component: () => import('@/view/portal/Layout.vue'),
        children: [
          {
            path: '',
            name: 'WebHome',
            meta: { title: '首页', client: true },
            component: () => import('@/view/portal/Home.vue')
          },
          {
            path: 'article/:slug',
            name: 'WebArticle',
            meta: { title: '文章', client: true },
            component: () => import('@/view/portal/ArticleDetail.vue')
          },
          {
            path: 'short-videos',
            name: 'WebShortVideoList',
            meta: { title: '短视频', client: true },
            component: () => import('@/view/portal/ShortVideoList.vue')
          },
          {
            path: 'color-blindness',
            name: 'WebColorBlindness',
            meta: { title: '色盲色弱', client: true },
            component: () => import('@/view/portal/ColorBlindness.vue')
          },
          {
            path: 'short-video/:slug',
            name: 'WebShortVideo',
            meta: { title: '短视频详情', client: true },
            component: () => import('@/view/portal/ShortVideoDetail.vue')
          },
          {
            path: 'contact',
            name: 'WebContact',
            meta: { title: '联系方式', client: true },
            component: () => import('@/view/portal/Contact.vue')
          },
          {
            path: 'office-tools',
            name: 'WebOfficeTools',
            meta: { title: '办公工具', client: true },
            component: () => import('@/view/portal/OfficeTools.vue')
          },
          {
            path: 'profile',
            redirect: { name: 'WebMember' }
          },
          {
            path: 'member',
            name: 'WebMember',
            meta: { title: '会员中心', client: true },
            component: () => import('@/view/portal/MemberCenter.vue')
          },
          {
            path: 'member/register',
            name: 'WebMemberRegister',
            meta: { title: '会员注册', client: true },
            component: () => import('@/view/portal/MemberRegister.vue')
          },
          {
            path: 'login',
            name: 'MemberEntryLogin',
            meta: { title: '会员登录', client: true },
            component: () => import('@/view/portal/MemberLogin.vue')
          }
        ]
      },
      {
        path: '/:catchAll(.*)',
        meta: { title: '404', client: true, closeTab: true },
        component: () => import('@/view/error/index.vue')
      }
    ]
  }

  return [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/init',
      name: 'Init',
      component: () => import('@/view/init/index.vue')
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/view/login/index.vue')
    },
    {
      path: '/scanUpload',
      name: 'ScanUpload',
      meta: {
        title: '扫码上传',
        client: true
      },
      component: () => import('@/view/example/upload/scanUpload.vue')
    },
    {
      path: '/:catchAll(.*)',
      meta: {
        closeTab: true
      },
      component: () => import('@/view/error/index.vue')
    }
  ]
}

const router = createRouter({
  history: createWebHashHistory(getHistoryBase()),
  routes: buildRoutes()
})

export default router
