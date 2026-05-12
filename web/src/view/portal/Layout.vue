<template>
  <div class="portal-root">
    <header class="portal-header">
      <div class="portal-inner header-flex">
        <router-link to="/" class="brand">内容站</router-link>
        <nav class="nav">
          <router-link to="/">首页</router-link>
          <router-link to="/member">会员中心</router-link>
          <router-link v-if="!token" class="muted" :to="{ name: 'MemberEntryLogin', query: { redirect: '/member' } }">
            会员登录
          </router-link>
        </nav>
      </div>
    </header>
    <div v-if="isHome" class="portal-hero-wrap portal-inner">
      <PortalHomeCarousel />
    </div>
    <main class="portal-main">
      <div class="portal-inner">
        <router-view />
      </div>
    </main>
    <footer class="portal-footer">
      <div class="portal-inner footer-inner">
        <div class="footer-row">
          <div class="footer-left">
            <span>©2008-2026 内容站 版权所有</span>
            <a
              class="footer-muted"
              href="https://beian.miit.gov.cn/"
              target="_blank"
              rel="noopener noreferrer"
            >湘ICP备2026010094号</a>
            
          </div>
          <nav class="footer-nav" aria-label="页脚链接">
            <template v-for="(item, idx) in footerNavLinks" :key="item.label">
              <span v-if="idx > 0" class="footer-bar">|</span>
              <a :href="item.href" class="footer-nav-link">{{ item.label }}</a>
            </template>
          </nav>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '@/pinia/modules/user'
  import PortalHomeCarousel from '@/view/portal/components/PortalHomeCarousel.vue'

  const route = useRoute()
  const userStore = useUserStore()
  const { token } = storeToRefs(userStore)

  const isHome = computed(() => route.name === 'WebHome')

  /** 页脚右侧导航（先写死，后续可接 CMS / 路由） */
  const footerNavLinks = [
    { label: '公司简介', href: '#' },
    { label: '联系方式', href: '#' },
    { label: '合作代理', href: '#' },
    { label: '隐私政策', href: '#' },
    { label: '使用协议', href: '#' },
    { label: '意见反馈', href: '#' }
  ]
</script>

<style scoped>
  .portal-root {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--portal-bg, #f5f5f5);
    color: #1a1a1a;
    /* 页面画布 #f5f5f5，内容块仍用 --portal-panel-bg */
    --portal-bg: #f5f5f5;
    --portal-panel-bg: #ffffff;
    --portal-muted-bg: #ffffff;
    --portal-radius: 12px;
    --portal-card-bg: #ffffff;
    --portal-text-secondary: #6b7280;
    --portal-text-body: #4b5563;
    --portal-text-meta: #9ca3af;
    --portal-link: #2563eb;
    --portal-hairline: #f3f4f6;
  }
  .portal-inner {
    max-width: 960px;
    margin: 0 auto;
    padding: 0 20px;
    width: 100%;
  }
  .portal-header {
    background: var(--portal-panel-bg, #ffffff);
    border-bottom: 1px solid var(--portal-hairline, #f3f4f6);
    position: sticky;
    top: 0;
    z-index: 10;
  }
  .header-flex {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
  }
  .brand {
    font-weight: 700;
    font-size: 1.1rem;
    color: #2563eb;
    text-decoration: none;
  }
  .nav {
    display: flex;
    align-items: center;
    gap: 1.25rem;
  }
  .nav a {
    color: #374151;
    text-decoration: none;
    font-size: 0.95rem;
  }
  .nav a.router-link-active:not(.muted) {
    color: #2563eb;
    font-weight: 600;
  }
  .nav a:hover {
    color: #1d4ed8;
  }
  .muted {
    opacity: 0.85;
  }
  .portal-main {
    flex: 1;
    padding: 28px 0 40px;
  }

  .portal-hero-wrap {
    padding-top: 12px;
    padding-bottom: 4px;
    box-sizing: border-box;
  }

  .portal-footer {
    border-top: 1px solid #e5e5e5;
    background: #f5f5f5;
    padding: 14px 0 18px;
  }

  .footer-inner {
    box-sizing: border-box;
  }

  .footer-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px 20px;
    font-size: 12px;
    line-height: 1.5;
    color: #333333;
  }

  .footer-left {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 14px;
    min-width: 0;
  }

  .footer-muted {
    color: #555555;
    text-decoration: none;
  }

  .footer-muted:hover {
    color: #2563eb;
    text-decoration: underline;
  }

  .footer-beian {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .footer-beian-icon {
    display: inline-block;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, #5a9fd4, #1a5276);
    flex-shrink: 0;
  }

  .footer-nav {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 0;
    min-width: 0;
  }

  .footer-bar {
    margin: 0 8px;
    color: #cccccc;
    user-select: none;
  }

  .footer-nav-link {
    color: #333333;
    text-decoration: none;
    white-space: nowrap;
  }

  .footer-nav-link:hover {
    color: #2563eb;
    text-decoration: underline;
  }

  @media (max-width: 720px) {
    .footer-row {
      flex-direction: column;
      align-items: flex-start;
    }
    .footer-nav {
      justify-content: flex-start;
    }
  }
</style>
