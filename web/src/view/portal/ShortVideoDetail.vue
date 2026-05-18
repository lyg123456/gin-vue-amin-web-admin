<template>
  <div class="sv-detail-page">
    <el-button text type="primary" class="back" @click="$router.push({ name: 'WebShortVideoList' })">
      ← 返回短视频列表
    </el-button>

    <el-skeleton v-if="loading" :rows="8" animated />
    <el-empty v-else-if="!video" description="短视频不存在或未发布" />
    <article v-else class="detail">
      <h1>{{ video.title }}</h1>
      <div class="meta">
        <span>{{ formatDate(video.publishedAt || video.UpdatedAt) }}</span>
        <span>播放 {{ video.viewCount ?? 0 }}</span>
        <span v-if="video.durationSec">时长 {{ video.durationSec }} 秒</span>
      </div>
      <p v-if="video.description" class="desc">{{ video.description }}</p>
      <div v-if="playUrl" class="player-box">
        <video :src="playUrl" controls playsinline class="player" />
      </div>
      <div v-if="scriptHtml" class="script-box">
        <h2>脚本 / 文案</h2>
        <div class="script-body" v-html="scriptHtml" />
      </div>
    </article>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { getPublishedShortVideoBySlug } from '@/api/contentShortVideo'
  import { formatDate } from '@/utils/format'
  import { getUrl } from '@/utils/image'
  import { Marked } from 'marked'

  const route = useRoute()
  const video = ref(null)
  const loading = ref(true)
  const marked = new Marked()

  const playUrl = computed(() => {
    const u = video.value?.videoUrl
    return u ? getUrl(u) : ''
  })

  const scriptHtml = computed(() => {
    const s = video.value?.script
    if (!s) return ''
    return marked.parse(s)
  })

  const load = async () => {
    loading.value = true
    try {
      const res = await getPublishedShortVideoBySlug(route.params.slug)
      if (res.code === 0) {
        video.value = res.data
      } else {
        video.value = null
      }
    } finally {
      loading.value = false
    }
  }

  onMounted(load)
  watch(() => route.params.slug, load)
</script>

<style scoped>
  .sv-detail-page { padding-bottom: 40px; }
  .back { margin-bottom: 12px; }
  .detail h1 { margin: 0 0 12px; font-size: 1.6rem; }
  .meta { color: #909399; font-size: 14px; display: flex; gap: 16px; flex-wrap: wrap; margin-bottom: 16px; }
  .desc { color: #606266; line-height: 1.6; margin-bottom: 20px; }
  .player-box { margin-bottom: 24px; border-radius: 12px; overflow: hidden; background: #000; }
  .player { width: 100%; max-height: 70vh; display: block; }
  .script-box h2 { font-size: 1.1rem; margin: 0 0 12px; }
  .script-body { line-height: 1.7; color: #303133; }
  .script-body :deep(p) { margin: 0.5em 0; }
</style>
