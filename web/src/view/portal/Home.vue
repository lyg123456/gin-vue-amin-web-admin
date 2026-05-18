<template>
  <div class="portal-home">
    <portal-pricing-tiers />

    <section class="portal-panel short-video-home" aria-label="短视频">
      <div class="sv-home-head">
        <h2 class="sv-home-title">短视频获客</h2>
        <router-link :to="{ name: 'WebShortVideoList' }" class="sv-home-more">查看更多 →</router-link>
      </div>
      <el-skeleton v-if="svLoading" :rows="3" animated />
      <el-empty v-else-if="!shortVideos.length" description="暂无已发布短视频" />
      <ul v-else class="sv-home-list">
        <li v-for="item in shortVideos" :key="item.ID" class="sv-home-row" @click="goShortVideo(item.slug)">
          <span class="d-bullet" aria-hidden="true" />
          <span class="d-link">{{ item.title }}</span>
          <span class="sv-dur">{{ item.durationSec || 0 }}s</span>
          <time class="d-date">{{ formatMonthDay(item.publishedAt || item.UpdatedAt) }}</time>
        </li>
      </ul>
    </section>
    <section class="portal-panel">
    <p class="intro">以下为管理端「内容 → 文章」中已发布内容，无需登录。</p>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        clearable
        placeholder="搜索标题或摘要"
        class="search-input"
        @keyup.enter="reload"
        @clear="reload"
      />
      <el-button type="primary" plain @click="reload">搜索</el-button>
    </div>

    <el-skeleton v-if="loading" :rows="8" animated class="skel" />
    <el-empty v-else-if="!list.length" description="暂无已发布文章，请在管理端「内容 → 文章」中编辑并发布。" />

    <div v-else class="columns">
      <section v-for="(col, idx) in columnData" :key="idx" class="d-card">
        <h2 class="d-card-title">{{ colTitles[idx] }}</h2>
        <div class="d-divider" />
        <ul class="d-list">
          <li v-for="item in col" :key="item.ID" class="d-row" @click="goArticle(item.slug)">
            <span class="d-bullet" aria-hidden="true" />
            <span class="d-link">{{ item.title }}</span>
            <time class="d-date">{{ formatMonthDay(item.publishedAt || item.UpdatedAt) }}</time>
          </li>
        </ul>
      </section>
    </div>

    <div v-if="total > pageSize" class="pager">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="onPage"
      />
    </div>
    </section>


    <section class="home-contact-strip" aria-label="联系方式">
      <div class="home-contact-inner">
        <div class="home-contact-text">
          <span class="home-contact-name">清风</span>
          <span class="home-contact-phone">19225501831</span>
          <router-link class="home-contact-more" :to="{ name: 'WebContact' }">更多联系方式</router-link>
        </div>
        <div class="home-contact-qr">
          <span class="home-contact-wechat-label">微信</span>
          <img src="/portal/wechat-qingfeng.png" alt="微信二维码 — 清风" width="88" height="88" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { usePublicArticleList } from '@/view/portal/composables/usePublicArticleList'
  import PortalPricingTiers from '@/view/portal/components/PortalPricingTiers.vue'
  import { getPublishedShortVideos } from '@/api/contentShortVideo'

  const colTitles = ['新闻动态', '程序发布', '文档教程']

  const {
    list,
    total,
    page,
    pageSize,
    keyword,
    loading,
    load,
    reload,
    onPage,
    goArticle
  } = usePublicArticleList({ initialPageSize: 10 })

  /** 当前页列表均分到 3 栏（轮询分配，不改变接口与分页逻辑） */
  const columnData = computed(() => {
    const cols = [[], [], []]
    const arr = list.value || []
    arr.forEach((item, i) => {
      cols[i % 3].push(item)
    })
    return cols
  })

  function formatMonthDay(time) {
    if (!time) return '--'
    const d = new Date(time)
    if (Number.isNaN(d.getTime())) return '--'
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${m}-${day}`
  }

  const router = useRouter()
  const shortVideos = ref([])
  const svLoading = ref(true)

  const loadShortVideos = async () => {
    svLoading.value = true
    try {
      const res = await getPublishedShortVideos({ page: 1, pageSize: 8 })
      if (res.code === 0) {
        shortVideos.value = res.data?.list || []
      }
    } finally {
      svLoading.value = false
    }
  }

  const goShortVideo = (slug) => {
    router.push({ name: 'WebShortVideo', params: { slug } })
  }

  onMounted(() => {
    load()
    loadShortVideos()
  })
</script>

<style scoped>
  .portal-home {
    display: flex;
    flex-direction: column;
    gap: 20px;
    min-height: 48vh;
  }

  .portal-panel {
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    border: none;
    padding: 24px;
  }

  .intro {
    margin: 0 0 16px;
    font-size: 0.9rem;
    color: var(--portal-text-secondary, #6b7280);
    line-height: 1.5;
  }

  .toolbar {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
    max-width: 520px;
    margin-bottom: 20px;
  }

  .search-input {
    flex: 1;
    min-width: 200px;
  }

  .skel {
    background: var(--portal-card-bg, #ffffff);
    padding: 16px;
    border-radius: var(--portal-radius, 12px);
  }

  .columns {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 14px;
    align-items: stretch;
  }

  .d-card {
    background: var(--portal-card-bg, #ffffff);
    border: none;
    border-radius: var(--portal-radius, 12px);
    box-sizing: border-box;
    padding: 0 0 12px;
    min-height: 280px;
    overflow: hidden;
  }

  .d-card-title {
    margin: 0;
    padding: 14px 16px 10px;
    text-align: center;
    font-size: 1.05rem;
    font-weight: 700;
    color: #1a1a1a;
    letter-spacing: 0.02em;
  }

  .d-divider {
    height: 1px;
    background: var(--portal-hairline, #f3f4f6);
    margin: 0 12px 8px;
  }

  .d-list {
    list-style: none;
    margin: 0;
    padding: 0 14px 8px;
  }

  .d-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 6px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--portal-text-body, #4b5563);
  }

  .d-bullet {
    width: 4px;
    height: 4px;
    flex-shrink: 0;
    background: #d1d5db;
    border-radius: 50%;
  }

  .d-link {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .d-date {
    flex-shrink: 0;
    font-size: 0.78rem;
    color: var(--portal-text-meta, #9ca3af);
    font-variant-numeric: tabular-nums;
  }

  .pager {
    margin-top: 28px;
    display: flex;
    justify-content: center;
  }

  .short-video-home {
    margin-top: 0;
  }
  .sv-home-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }
  .sv-home-title {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--portal-text, #303133);
  }
  .sv-home-more {
    font-size: 14px;
    color: var(--el-color-primary);
    text-decoration: none;
  }
  .sv-home-more:hover {
    text-decoration: underline;
  }
  .sv-home-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .sv-home-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    gap: 10px 12px;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--portal-border, #ebeef5);
    cursor: pointer;
  }
  .sv-home-row:last-child {
    border-bottom: none;
  }
  .sv-home-row:hover .d-link {
    color: var(--el-color-primary);
  }
  .sv-dur {
    font-size: 12px;
    color: #909399;
  }
  .home-contact-strip {
    margin-top: 4px;
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    padding: 16px 20px;
    box-sizing: border-box;
  }

  .home-contact-inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
  }

  .home-contact-text {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 0.95rem;
    color: var(--portal-text-body, #4b5563);
  }

  .home-contact-name {
    font-weight: 700;
    color: #1a1a1a;
    font-size: 1rem;
  }

  .home-contact-phone {
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }

  .home-contact-more {
    margin-top: 4px;
    font-size: 0.85rem;
    color: var(--portal-link, #2563eb);
    text-decoration: none;
  }

  .home-contact-more:hover {
    text-decoration: underline;
  }

  .home-contact-qr {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-shrink: 0;
  }

  .home-contact-wechat-label {
    font-size: 0.85rem;
    color: var(--portal-text-secondary, #6b7280);
  }

  .home-contact-qr img {
    width: 88px;
    height: 88px;
    object-fit: contain;
    border-radius: 6px;
    box-shadow: 0 2px 10px rgba(15, 23, 42, 0.08);
  }

  @media (max-width: 960px) {
    .columns {
      grid-template-columns: 1fr;
    }
  }
</style>
