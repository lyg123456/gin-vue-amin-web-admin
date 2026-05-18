<template>
  <div class="sv-list-page">
    <h1 class="page-title">短视频</h1>
    <p class="intro">精选已发布短视频，点击卡片进入详情播放。</p>

    <el-skeleton v-if="loading" :rows="6" animated />
    <el-empty v-else-if="!list.length" description="暂无短视频" />

    <div v-else class="grid">
      <article
        v-for="item in list"
        :key="item.ID"
        class="card"
        @click="goDetail(item.slug)"
      >
        <div class="thumb">
          <img v-if="coverSrc(item)" :src="coverSrc(item)" :alt="item.title" />
          <div v-else class="thumb-placeholder">短视频</div>
          <span class="duration">{{ item.durationSec || 0 }}s</span>
        </div>
        <h2 class="card-title">{{ item.title }}</h2>
        <p v-if="item.description" class="card-desc">{{ item.description }}</p>
        <div class="card-meta">
          <span>{{ formatDate(item.publishedAt || item.UpdatedAt) }}</span>
          <span>播放 {{ item.viewCount ?? 0 }}</span>
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
  import { getPublishedShortVideos } from '@/api/contentShortVideo'
  import { formatDate } from '@/utils/format'
  import { getUrl } from '@/utils/image'

  const router = useRouter()
  const list = ref([])
  const loading = ref(true)
  const page = ref(1)
  const pageSize = ref(12)
  const total = ref(0)

  const coverSrc = (item) => {
    const u = item.coverImage || ''
    const first = String(u).split(',')[0]?.trim()
    return first ? getUrl(first) : ''
  }

  const load = async () => {
    loading.value = true
    try {
      const res = await getPublishedShortVideos({ page: page.value, pageSize: pageSize.value })
      if (res.code === 0) {
        list.value = res.data.list || []
        total.value = res.data.total
      }
    } finally {
      loading.value = false
    }
  }

  const onPage = (p) => {
    page.value = p
    load()
  }

  const goDetail = (slug) => {
    router.push({ name: 'WebShortVideo', params: { slug } })
  }

  onMounted(load)
</script>

<style scoped>
  .sv-list-page { padding: 8px 0 32px; }
  .page-title { margin: 0 0 8px; font-size: 1.5rem; }
  .intro { color: #606266; margin: 0 0 20px; font-size: 14px; }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 20px;
  }
  .card {
    cursor: pointer;
    border-radius: 12px;
    overflow: hidden;
    border: 1px solid #ebeef5;
    background: #fff;
    transition: box-shadow 0.2s;
  }
  .card:hover { box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08); }
  .thumb {
    position: relative;
    aspect-ratio: 16/9;
    background: #f5f7fa;
  }
  .thumb img { width: 100%; height: 100%; object-fit: cover; }
  .thumb-placeholder {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
  }
  .duration {
    position: absolute;
    right: 8px;
    bottom: 8px;
    background: rgba(0, 0, 0, 0.65);
    color: #fff;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 4px;
  }
  .card-title {
    margin: 12px 14px 6px;
    font-size: 15px;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .card-desc {
    margin: 0 14px 8px;
    font-size: 13px;
    color: #909399;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .card-meta {
    padding: 0 14px 14px;
    font-size: 12px;
    color: #909399;
    display: flex;
    justify-content: space-between;
  }
  .pager { margin-top: 24px; display: flex; justify-content: center; }
</style>
