<template>
  <div class="article-page">
    <el-button text type="primary" class="back" @click="$router.push({ name: 'WebHome' })">
      ← 返回列表
    </el-button>

    <el-skeleton v-if="loading" :rows="10" animated />
    <el-empty v-else-if="!article" description="文章不存在或未发布" />
    <article v-else class="article">
      <header class="article-head">
        <h1>{{ article.title }}</h1>
        <div class="sub">
          <span>{{ formatDate(article.publishedAt || article.UpdatedAt) }}</span>
          <span>阅读 {{ article.viewCount ?? 0 }}</span>
        </div>
        <p v-if="article.summary" class="summary">{{ article.summary }}</p>
        <div
          v-if="coverUrls.length"
          class="covers"
          :class="{
            'covers--multi': coverUrls.length > 1,
            'covers--single': coverUrls.length === 1
          }"
        >
          <div
            v-for="(u, i) in coverUrls"
            :key="i"
            :class="coverUrls.length === 1 ? 'cover-wrap-hero' : 'cover-wrap-cell'"
          >
            <el-image
              :src="coverSrc(u)"
              :preview-src-list="previewList"
              :initial-index="i"
              :fit="coverUrls.length === 1 ? 'contain' : 'cover'"
              preview-teleported
              hide-on-click-modal
              class="cover-el"
            />
          </div>
        </div>
      </header>
      <div v-if="article.contentType === 'html'" class="body html-body" v-html="article.content" />
      <div v-else class="body md-body" v-html="htmlFromMd" />
    </article>
  </div>
</template>

<script setup>
  import { Marked } from 'marked'
  import { markedHighlight } from 'marked-highlight'
  import hljs from 'highlight.js'
  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { getPublicArticleBySlug } from '@/api/publicArticle'
  import config from '@/core/config'
  import { formatDate } from '@/utils/format'
  import { getUrl, coverImageUrls } from '@/utils/image'

  const route = useRoute()
  const article = ref(null)
  const loading = ref(true)

  onMounted(() => {
    import('highlight.js/styles/atom-one-light.css')
  })

  const coverSrc = (url) => getUrl(url)

  const coverUrls = computed(() => coverImageUrls(article.value?.coverImage))

  const previewList = computed(() => coverUrls.value.map((u) => coverSrc(u)))

  const marked = new Marked(
    markedHighlight({
      langPrefix: 'hljs language-',
      highlight(code, lang) {
        const language = hljs.getLanguage(lang) ? lang : 'plaintext'
        return hljs.highlight(code, { language }).value
      }
    })
  )

  const htmlFromMd = computed(() => {
    const raw = article.value?.content
    if (!raw) return ''
    try {
      return marked.parse(raw)
    } catch {
      return '<p>正文解析失败</p>'
    }
  })

  const load = async () => {
    const slug = route.params.slug
    if (!slug) {
      article.value = null
      loading.value = false
      return
    }
    loading.value = true
    article.value = null
    try {
      const res = await getPublicArticleBySlug(String(slug))
      if (res.code === 0) {
        article.value = res.data
        const t = res.data?.seoTitle || res.data?.title
        if (t) {
          document.title = `${t} - ${config.appName}`
        }
      }
    } finally {
      loading.value = false
    }
  }

  watch(
    () => route.params.slug,
    () => load(),
    { immediate: true }
  )
</script>

<style scoped>
  .back {
    margin-bottom: 16px;
    padding-left: 0;
  }
  .article {
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    border: none;
    padding: 28px 24px 40px;
  }
  .article-head h1 {
    margin: 0 0 12px;
    font-size: 1.75rem;
    line-height: 1.3;
    font-weight: 700;
  }
  .sub {
    display: flex;
    gap: 16px;
    font-size: 0.85rem;
    color: #9ca3af;
    margin-bottom: 16px;
  }
  .summary {
    margin: 0 0 20px;
    color: #4b5563;
    line-height: 1.6;
    font-size: 1rem;
  }
  .covers {
    margin-bottom: 24px;
  }
  /* 单图：定高画幅 + contain，避免竖图/Logo 被 cover 裁切 */
  .covers--single .cover-wrap-hero {
    width: 100%;
    height: max(220px, min(52vh, 480px));
    border-radius: var(--portal-radius, 12px);
    overflow: hidden;
    border: none;
    background: #ffffff;
  }
  .covers--single .cover-wrap-hero .cover-el {
    width: 100%;
    height: 100%;
    display: block;
  }
  .covers--single .cover-wrap-hero :deep(.el-image),
  .covers--single .cover-wrap-hero :deep(.el-image__wrapper) {
    width: 100% !important;
    height: 100% !important;
  }
  .covers--single .cover-wrap-hero :deep(.el-image__inner) {
    cursor: zoom-in;
    object-position: center center;
  }
  .covers--multi {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(2, 1fr);
  }
  @media (min-width: 720px) {
    .covers--multi {
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    }
  }
  .cover-wrap-cell {
    aspect-ratio: 4 / 3;
    border-radius: 10px;
    overflow: hidden;
    border: none;
    box-shadow: 0 2px 10px rgba(15, 23, 42, 0.05);
  }
  .cover-wrap-cell .cover-el {
    width: 100%;
    height: 100%;
    display: block;
  }
  .cover-wrap-cell :deep(.el-image__inner) {
    width: 100%;
    height: 100%;
    object-fit: cover;
    cursor: zoom-in;
  }
  .body {
    font-size: 1rem;
    line-height: 1.75;
    color: var(--portal-text-body, #4b5563);
  }
</style>

<style>
  /* 正文 Markdown 区域（非 scoped） */
  .portal-root .md-body :deep(h1),
  .portal-root .md-body :deep(h2),
  .portal-root .md-body :deep(h3) {
    margin-top: 1.25em;
    margin-bottom: 0.5em;
    font-weight: 600;
  }
  .portal-root .md-body :deep(p) {
    margin: 0.75em 0;
  }
  .portal-root .md-body :deep(pre) {
    background: #f3f4f6;
    padding: 14px 16px;
    border-radius: 8px;
    overflow-x: auto;
    font-size: 0.9rem;
  }
  .portal-root .md-body :deep(code) {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .portal-root .html-body :deep(img) {
    max-width: 100%;
    height: auto;
  }
</style>
