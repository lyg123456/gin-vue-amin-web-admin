<template>
  <div class="home">
    <section class="hero">
      <h1>最新文章</h1>
      <p class="lead">以下为后台已发布内容，点击卡片阅读全文。</p>
      <div class="toolbar">
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
    </section>

    <el-skeleton v-if="loading" :rows="6" animated />
    <el-empty v-else-if="!list.length" description="暂无已发布文章" />
    <div v-else class="grid">
      <article v-for="item in list" :key="item.ID" class="card" @click="goArticle(item.slug)">
        <div v-if="item.coverImage" class="cover-wrap">
          <img :src="coverSrc(item.coverImage)" alt="" class="cover" />
        </div>
        <div class="card-body">
          <h2 class="title">{{ item.title }}</h2>
          <p class="summary">{{ item.summary || '暂无摘要' }}</p>
          <div class="meta">
            <span>{{ formatDate(item.publishedAt || item.UpdatedAt) }}</span>
            <span>阅读 {{ item.viewCount ?? 0 }}</span>
          </div>
        </div>
      </article>
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
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { getPublicArticleList } from '@/api/publicArticle'
  import { formatDate } from '@/utils/format'
  import { getUrl } from '@/utils/image'

  const router = useRouter()
  const list = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const keyword = ref('')
  const loading = ref(true)

  const coverSrc = (url) => getUrl(url)

  const load = async () => {
    loading.value = true
    try {
      const res = await getPublicArticleList({
        page: page.value,
        pageSize: pageSize.value,
        keyword: keyword.value || undefined
      })
      if (res.code === 0) {
        list.value = res.data.list || []
        total.value = res.data.total || 0
        page.value = res.data.page || 1
        pageSize.value = res.data.pageSize || 10
      }
    } finally {
      loading.value = false
    }
  }

  const reload = () => {
    page.value = 1
    load()
  }

  const onPage = (p) => {
    page.value = p
    load()
  }

  const goArticle = (slug) => {
    router.push({ name: 'WebArticle', params: { slug } })
  }

  onMounted(load)
</script>

<style scoped>
  .hero {
    margin-bottom: 28px;
  }
  .hero h1 {
    margin: 0 0 8px;
    font-size: 1.75rem;
    font-weight: 700;
  }
  .lead {
    margin: 0 0 16px;
    color: #6b7280;
    font-size: 0.95rem;
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
    gap: 16px;
  }
  .card {
    display: flex;
    gap: 16px;
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
    width: 200px;
    min-height: 120px;
    flex-shrink: 0;
    background: #e5e7eb;
  }
  .cover {
    width: 100%;
    height: 100%;
    min-height: 120px;
    object-fit: cover;
    display: block;
  }
  .card-body {
    padding: 16px 16px 16px 0;
    flex: 1;
    min-width: 0;
  }
  .title {
    margin: 0 0 8px;
    font-size: 1.15rem;
    font-weight: 600;
    line-height: 1.35;
  }
  .summary {
    margin: 0 0 12px;
    font-size: 0.9rem;
    color: #4b5563;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .meta {
    display: flex;
    gap: 16px;
    font-size: 0.8rem;
    color: #9ca3af;
  }
  .pager {
    margin-top: 28px;
    display: flex;
    justify-content: center;
  }
  @media (max-width: 640px) {
    .card {
      flex-direction: column;
    }
    .cover-wrap {
      width: 100%;
      min-height: 160px;
    }
    .card-body {
      padding: 0 16px 16px;
    }
  }
</style>
