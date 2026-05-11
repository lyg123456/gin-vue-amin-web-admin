<template>
  <div class="portal-home">
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
  </div>
</template>

<script setup>
  import { computed, onMounted } from 'vue'
  import { usePublicArticleList } from '@/view/portal/composables/usePublicArticleList'

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

  onMounted(load)
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

  @media (max-width: 960px) {
    .columns {
      grid-template-columns: 1fr;
    }
  }
</style>
