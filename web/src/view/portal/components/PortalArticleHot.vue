<template>
  <section class="feed">
    <div class="feed-head">
      <h2 class="feed-title">{{ title }}</h2>
      <p v-if="subtitle" class="feed-sub">{{ subtitle }}</p>
       <!-- 👉 这里显示页面访问量 -->
      
      <div v-if="showSearch" class="toolbar">
        <el-input
          v-model="keyword"
          clearable
          placeholder="搜索标题或摘要"
          class="search"
          @keyup.enter="reload"
          @clear="reload"
        />
        <el-button type="primary" @click="reload">搜索</el-button>
      </div>
    </div>

    <el-skeleton v-if="loading" :rows="4" animated />
    <el-empty v-else-if="!list.length" description="暂无已发布文章，请在管理端「内容 → 文章」中编辑并发布。" />
    <div v-else class="grid">
      <article v-for="item in list" :key="item.ID" class="card" @click="goArticle(item.slug)">
        <div v-if="item.coverImage" class="cover-wrap">
          <img :src="coverSrc(item.coverImage)" alt="" class="cover" />
        </div>
        <div class="card-body">
          <h3 class="title">{{ item.title }}</h3>
          <p class="summary">{{ item.summary || '暂无摘要' }}</p>
          <div class="meta">
            <span>{{ formatDate(item.publishedAt || item.UpdatedAt) }}</span>
            <span>阅读 {{ item.viewCount ?? 0 }}</span>
          </div>
        </div>
      </article>
    </div>

    <div v-if="showPager && total > pageSize" class="pager">
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
</template>

<script setup>
 
  import { onMounted, ref } from 'vue'
  import { usePublicArticleList } from '@/view/portal/composables/usePublicArticleList'
  import { formatDate } from '@/utils/format'
   import axios from 'axios'

  const props = defineProps({
    title: { type: String, default: '最新文章' },
    subtitle: { type: String, default: '' },
    showSearch: { type: Boolean, default: false },
    showPager: { type: Boolean, default: true },
    pageSize: { type: Number, default: 10 }
  })

  const {
    list,
    total,
    page,
    pageSize,
    keyword,
    loading,
    coverSrc,
    load,
    reload,
    onPage,
    goArticle
  } = usePublicArticleList({ initialPageSize: props.pageSize })
  const visit = ref({ total: 0, today: 0 })
  const loadVisitStats = async () => {
  try {
    // ✅ 完全和你列表一样格式！/api/public/xxx
    const res = await axios.get('/api/public/web/stats')
    
    console.log("✅ 访问量成功：", res.data)
    
    if (res.data.code === 0) {
      visit.value.total = res.data.data.total
      visit.value.today = res.data.data.today
    }
  } catch (e) {
    console.log("❌ 失败：", e)
  }
}

  onMounted(() => {
    load()
    loadVisitStats() // 页面打开 → 自动统计
  })
</script>

<style scoped>
  .feed-head {
    margin-bottom: 20px;
  }
  .feed-title {
    margin: 0 0 8px;
    font-size: 1.35rem;
    font-weight: 700;
  }
  .feed-sub {
    margin: 0 0 16px;
    color: #6b7280;
    font-size: 0.9rem;
    line-height: 1.5;
  }
  .toolbar {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    max-width: 480px;
  }
  .search {
    flex: 1;
    min-width: 200px;
  }
  .grid {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .card {
    display: flex;
    gap: 14px;
    background: #fff;
    border-radius: 12px;
    border: 1px solid #e8eaed;
    overflow: hidden;
    cursor: pointer;
    transition: box-shadow 0.2s, border-color 0.2s;
  }
  .card:hover {
    border-color: #bfdbfe;
    box-shadow: 0 8px 24px rgba(37, 99, 235, 0.08);
  }
  .cover-wrap {
    width: 180px;
    min-height: 110px;
    flex-shrink: 0;
    background: #e5e7eb;
  }
  .cover {
    width: 100%;
    height: 100%;
    min-height: 110px;
    object-fit: cover;
    display: block;
  }
  .card-body {
    padding: 14px 14px 14px 0;
    flex: 1;
    min-width: 0;
  }
  .title {
    margin: 0 0 6px;
    font-size: 1.05rem;
    font-weight: 600;
    line-height: 1.35;
  }
  .summary {
    margin: 0 0 10px;
    font-size: 0.88rem;
    color: #4b5563;
    line-height: 1.45;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .meta {
    display: flex;
    gap: 14px;
    font-size: 0.78rem;
    color: #9ca3af;
  }
  .pager {
    margin-top: 20px;
    display: flex;
    justify-content: center;
  }
  @media (max-width: 640px) {
    .card {
      flex-direction: column;
    }
    .cover-wrap {
      width: 100%;
      min-height: 150px;
    }
    .card-body {
      padding: 0 14px 14px;
    }
  }
  .visit-info {
    font-size: 0.85rem;
    color: #3b82f6;
    margin-bottom: 14px;
    background: #f0f7ff;
    padding: 6px 12px;
    border-radius: 6px;
    display: inline-block;
  }
</style>
