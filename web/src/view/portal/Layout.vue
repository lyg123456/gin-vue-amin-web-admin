<template>

  <div class="portal-root">

    <header class="portal-header">

      <div class="portal-inner header-flex">

        <router-link to="/" class="brand">

          <span class="brand-main">dnf1688</span>

          <span class="brand-slogan">在线工具</span>

        </router-link>

        <nav class="nav" aria-label="主导航">

          <router-link to="/">首页</router-link>

          <router-link :to="{ name: 'WebShortVideoList' }">短视频</router-link>
          <router-link :to="{ name: 'WebColorBlindness' }">色盲色弱</router-link>

          <div
            class="nav-dropdown"
            @mouseenter="videoCrawlOpen = true"
            @mouseleave="videoCrawlOpen = false"
          >
            <router-link
              :to="videoCrawlDefaultLink"
              class="nav-dropdown-trigger"
              :class="{ 'is-active': isVideoCrawlNavActive }"
            >
              抖音垂类抓取
              <span class="nav-caret" aria-hidden="true">▾</span>
            </router-link>
            <div v-show="videoCrawlOpen" class="nav-dropdown-menu" role="menu">
              <router-link
                v-for="item in videoCrawlNavItems"
                :key="item.id"
                :to="{ name: 'WebOfficeTools', query: { tool: item.tool } }"
                class="nav-dropdown-item"
                role="menuitem"
                @click="videoCrawlOpen = false"
              >
                {{ item.label }}
              </router-link>
            </div>
          </div>

          <router-link
            :to="{ name: 'WebOfficeTools' }"
            active-class=""
            exact-active-class=""
            :class="{ 'router-link-active': isOfficeToolsNavActive }"
          >
            在线工具
          </router-link>

          <router-link :to="{ name: 'WebContact' }">联系方式</router-link>

          <router-link to="/member">会员中心</router-link>

          <router-link v-if="!token" class="nav-login" :to="{ name: 'MemberEntryLogin', query: { redirect: '/member' } }">

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

            <span>©2008-2026 dnf1688 版权所有</span>

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

              <router-link v-if="item.route" :to="item.route" class="footer-nav-link">{{ item.label }}</router-link>

              <a v-else :href="item.href" class="footer-nav-link">{{ item.label }}</a>

            </template>

          </nav>

        </div>

        <p class="footer-site-notice">{{ siteFooterNotice }}</p>

      </div>

    </footer>

  </div>

</template>



<script setup>

  import { computed, ref } from 'vue'

  import { useRoute } from 'vue-router'
  import {
    DEFAULT_VIDEO_CRAWL_TOOL,
    VIDEO_CRAWL_NAV_ITEMS,
    VIDEO_CRAWL_TOOL_IDS
  } from '@/constants/videoCrawlNav'

  import { storeToRefs } from 'pinia'

  import { useUserStore } from '@/pinia/modules/user'

  import PortalHomeCarousel from '@/view/portal/components/PortalHomeCarousel.vue'
  import { PORTAL_SITE_FOOTER_NOTICE } from '@/constants/portalDisclaimer'

  const siteFooterNotice = PORTAL_SITE_FOOTER_NOTICE



  const route = useRoute()

  const userStore = useUserStore()

  const { token } = storeToRefs(userStore)



  const isHome = computed(() => route.name === 'WebHome')

  const videoCrawlNavItems = VIDEO_CRAWL_NAV_ITEMS
  const videoCrawlOpen = ref(false)
  const videoCrawlDefaultLink = {
    name: 'WebOfficeTools',
    query: { tool: DEFAULT_VIDEO_CRAWL_TOOL }
  }
  const isVideoCrawlNavActive = computed(
    () =>
      route.name === 'WebOfficeTools' &&
      VIDEO_CRAWL_TOOL_IDS.has(String(route.query.tool || ''))
  )
  const isOfficeToolsNavActive = computed(
    () =>
      route.name === 'WebOfficeTools' &&
      !VIDEO_CRAWL_TOOL_IDS.has(String(route.query.tool || ''))
  )



  const footerNavLinks = computed(() => {

    const contact = { label: '联系方式', route: { name: 'WebContact' } }

    const office = { label: '在线工具', route: { name: 'WebOfficeTools' } }

    const colorBlind = { label: '色盲色弱', route: { name: 'WebColorBlindness' } }

    const rest = [

      { label: '公司简介', href: '#' },

      { label: '隐私政策', href: '#' },

      { label: '意见反馈', href: '#' }

    ]

    return [contact, colorBlind, office, ...rest]

  })

</script>



<style lang="scss">

  @use '@/style/portal-theme.scss';

</style>



<style scoped lang="scss">

  .portal-root {

    min-height: 100vh;

    display: flex;

    flex-direction: column;

    background: var(--portal-bg, #f0f2f5);

    color: #333;

    --portal-bg: #f0f2f5;

    --portal-panel-bg: #ffffff;

    --portal-muted-bg: #ffffff;

    --portal-radius: 4px;

    --portal-card-bg: #ffffff;

    --portal-text-secondary: #666;

    --portal-text-body: #444;

    --portal-text-meta: #888;

    --portal-link: var(--portal-brand, #1a73e8);

    --portal-hairline: #e8e8e8;

  }



  .portal-header {

    background: var(--portal-header-bg, #fff);

    border-bottom: 1px solid var(--portal-header-border, #e0e0e0);

    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);

    position: sticky;

    top: 0;

    z-index: 100;

  }



  .header-flex {

    display: flex;

    align-items: center;

    justify-content: space-between;

    height: 52px;

    gap: 16px;

  }



  .brand {

    display: flex;

    align-items: baseline;

    gap: 8px;

    text-decoration: none;

    flex-shrink: 0;

  }



  .brand-main {

    font-size: 22px;

    font-weight: 700;

    color: var(--portal-brand, #1a73e8);

    letter-spacing: -0.02em;

    line-height: 1.2;

  }



  .brand-slogan {

    font-size: 13px;

    color: #888;

    font-weight: normal;

    padding-left: 8px;

    border-left: 1px solid #ddd;

  }



  .nav {

    display: flex;

    align-items: center;

    flex-wrap: wrap;

    gap: 0;

  }



  .nav a {

    color: #333;

    text-decoration: none;

    font-size: var(--portal-font-nav, 15px);

    padding: 16px 14px;

    line-height: 1;

    white-space: nowrap;

    border-bottom: 2px solid transparent;

    margin-bottom: -1px;

    transition: color 0.15s, border-color 0.15s;

  }



  .nav a:hover {

    color: var(--portal-brand, #1a73e8);

  }



  .nav a.router-link-active:not(.nav-login) {

    color: var(--portal-brand, #1a73e8);

    font-weight: 600;

    border-bottom-color: var(--portal-accent, #ff6600);

  }



  .nav-login {

    color: #666 !important;

    font-size: 14px !important;

  }



  .nav-login:hover {

    color: var(--portal-brand, #1a73e8) !important;

  }

  .nav-dropdown {
    position: relative;
    display: inline-flex;
    align-items: stretch;
  }

  .nav-dropdown-trigger {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .nav-dropdown-trigger.is-active {
    color: var(--portal-brand, #1a73e8);
    font-weight: 600;
    border-bottom-color: var(--portal-accent, #ff6600);
  }

  .nav-caret {
    font-size: 10px;
    opacity: 0.75;
    line-height: 1;
  }

  .nav-dropdown-menu {
    position: absolute;
    top: 100%;
    left: 0;
    min-width: 168px;
    padding: 6px 0;
    background: #fff;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
    z-index: 200;
  }

  .nav-dropdown-item {
    display: block;
    padding: 10px 16px;
    font-size: 14px;
    color: #333;
    text-decoration: none;
    white-space: nowrap;
    border-bottom: none !important;
    margin-bottom: 0 !important;
    line-height: 1.4;
  }

  .nav-dropdown-item:hover {
    background: #f5f8ff;
    color: var(--portal-brand, #1a73e8);
  }

  .nav-dropdown-item.router-link-active {
    color: var(--portal-brand, #1a73e8);
    font-weight: 600;
    background: #f0f6ff;
    border-bottom: none !important;
  }

  .portal-main {

    flex: 1;

    padding: 12px 0 28px;

  }



  .portal-hero-wrap {

    padding-top: 10px;

    padding-bottom: 4px;

    box-sizing: border-box;

  }



  .portal-footer {

    border-top: 1px solid #e0e0e0;

    background: #f5f5f5;

    padding: 14px 0 18px;

  }



  .footer-row {

    display: flex;

    justify-content: space-between;

    align-items: center;

    flex-wrap: wrap;

    gap: 12px 20px;

    font-size: 13px;

    line-height: 1.5;

    color: #666;

  }



  .footer-left {

    display: flex;

    flex-wrap: wrap;

    align-items: center;

    gap: 10px 14px;

    min-width: 0;

  }



  .footer-muted {

    color: #888;

    text-decoration: none;

  }



  .footer-muted:hover {

    color: var(--portal-brand, #1a73e8);

    text-decoration: underline;

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

    color: #ccc;

    user-select: none;

  }



  .footer-nav-link {

    color: #666;

    text-decoration: none;

    white-space: nowrap;

    font-size: 13px;

  }



  .footer-nav-link:hover {

    color: var(--portal-brand, #1a73e8);

    text-decoration: underline;

  }

  .footer-site-notice {
    width: 100%;
    margin: 10px 0 0;
    padding-top: 10px;
    border-top: 1px dashed #ddd;
    font-size: 12px;
    line-height: 1.65;
    color: #888;
    text-align: center;
  }

  @media (max-width: 768px) {

    .header-flex {

      flex-direction: column;

      align-items: flex-start;

      height: auto;

      padding: 10px 0;

    }



    .nav {

      width: 100%;

      overflow-x: auto;

    }



    .nav a {

      padding: 10px 12px;

      font-size: 14px;

    }



    .brand-slogan {

      display: none;

    }



    .footer-row {

      flex-direction: column;

      align-items: flex-start;

    }



    .footer-nav {

      justify-content: flex-start;

    }

  }

</style>

