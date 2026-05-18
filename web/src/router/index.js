import { createRouter, createWebHashHistory } from 'vue-router'

/**
 * 与部署路径一致：管理端 /admin/ ，用户端 /web/
 * 例：http://127.0.0.1:8080/admin/#/login 、http://127.0.0.1:8080/web/#/login
 */
export function getHistoryBase() {
  if (typeof window === 'undefined') return '/admin/'
  const p = window.location.pathname || ''
  if (p === '/web' || p.startsWith('/web/')) return '/web/'
  if (p === '/admin' || p.startsWith('/admin/')) return '/admin/'
  return '/admin/'
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
