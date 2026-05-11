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
    <main class="portal-main">
      <div class="portal-inner">
        <router-view />
      </div>
    </main>
    <footer class="portal-footer">
      <div class="portal-inner footer-text">
        <span>本站文章由管理后台发布 </span>
       <!-- <span class="sep">|</span>
         <a class="admin-link" :href="adminLoginHref">管理员入口</a>-->
      </div>
      
    </footer>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '@/pinia/modules/user'

  const userStore = useUserStore()
  const { token } = storeToRefs(userStore)

  const adminLoginHref = computed(() => {
    const o = typeof window !== 'undefined' ? window.location.origin : ''
    return `${o}/admin/#/login`
  })
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
  .portal-footer {
    border-top: 1px solid var(--portal-hairline, #f3f4f6);
    background: var(--portal-panel-bg, #ffffff);
    padding: 16px 0;
  }
  .footer-text {
    text-align: center;
    font-size: 0.8rem;
    color: var(--portal-text-secondary, #6b7280);
  }
  .sep {
    margin: 0 10px;
    color: #d1d5db;
  }
  .admin-link {
    color: #9ca3af;
    text-decoration: none;
  }
  .admin-link:hover {
    color: #6b7280;
    text-decoration: underline;
  }
  a.admin-link {
    cursor: pointer;
  }
</style>
