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
  // GFM 默认已开；显式关闭 breaks，避免「单换行」变成 <br> 与后台录入习惯不一致
  marked.setOptions({ gfm: true, breaks: false })

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
    word-break: break-word;
    overflow-wrap: anywhere;
  }
</style>

<style>
  /*
   * 正文样式必须写在「非 scoped」块里，且不能使用 :deep（:deep 仅对 scoped 生效，否则选择器无效，正文会像「没样式」一样错乱）
   * 同时覆盖 Markdown 与 HTML 两种 contentType
   */
  .portal-root .article .md-body,
  .portal-root .article .html-body {
    font-size: 1rem;
    line-height: 1.75;
    color: #374151;
    overflow-x: auto;
  }

  .portal-root .article .md-body > *:first-child,
  .portal-root .article .html-body > *:first-child {
    margin-top: 0;
  }

  .portal-root .article .md-body > *:last-child,
  .portal-root .article .html-body > *:last-child {
    margin-bottom: 0;
  }

  .portal-root .article .md-body h1,
  .portal-root .article .md-body h2,
  .portal-root .article .md-body h3,
  .portal-root .article .md-body h4,
  .portal-root .article .md-body h5,
  .portal-root .article .md-body h6,
  .portal-root .article .html-body h1,
  .portal-root .article .html-body h2,
  .portal-root .article .html-body h3,
  .portal-root .article .html-body h4,
  .portal-root .article .html-body h5,
  .portal-root .article .html-body h6 {
    margin-top: 1.15em;
    margin-bottom: 0.45em;
    font-weight: 600;
    line-height: 1.35;
    color: #111827;
  }

  .portal-root .article .md-body h1:first-child,
  .portal-root .article .html-body h1:first-child {
    margin-top: 0;
  }

  .portal-root .article .md-body p,
  .portal-root .article .html-body p {
    margin: 0.75em 0;
  }

  .portal-root .article .md-body ul,
  .portal-root .article .md-body ol,
  .portal-root .article .html-body ul,
  .portal-root .article .html-body ol {
    margin: 0.75em 0;
    padding-left: 1.5em;
  }

  .portal-root .article .md-body li,
  .portal-root .article .html-body li {
    margin: 0.35em 0;
  }

  .portal-root .article .md-body li > p,
  .portal-root .article .html-body li > p {
    margin: 0.25em 0;
  }

  .portal-root .article .md-body blockquote,
  .portal-root .article .html-body blockquote {
    margin: 1em 0;
    padding: 0.35em 0 0.35em 1em;
    border-left: 4px solid #e5e7eb;
    color: #6b7280;
    background: #f9fafb;
  }

  .portal-root .article .md-body hr,
  .portal-root .article .html-body hr {
    margin: 1.5em 0;
    border: none;
    border-top: 1px solid #e5e7eb;
  }

  .portal-root .article .md-body a,
  .portal-root .article .html-body a {
    color: #2563eb;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .portal-root .article .md-body a:hover,
  .portal-root .article .html-body a:hover {
    color: #1d4ed8;
  }

  .portal-root .article .md-body img,
  .portal-root .article .html-body img,
  .portal-root .article .md-body video,
  .portal-root .article .html-body video {
    max-width: 100%;
    height: auto;
    vertical-align: middle;
  }

  .portal-root .article .md-body table,
  .portal-root .article .html-body table {
    width: 100%;
    max-width: 100%;
    border-collapse: collapse;
    margin: 1em 0;
    font-size: 0.95rem;
    border: 1px solid #e5e7eb;
  }

  .portal-root .article .md-body th,
  .portal-root .article .md-body td,
  .portal-root .article .html-body th,
  .portal-root .article .html-body td {
    border: 1px solid #e5e7eb;
    padding: 8px 12px;
    text-align: left;
    vertical-align: top;
    word-break: break-word;
  }

  .portal-root .article .md-body th,
  .portal-root .article .html-body th {
    background: #f3f4f6;
    font-weight: 600;
  }

  .portal-root .article .md-body pre,
  .portal-root .article .html-body pre {
    margin: 1em 0;
    padding: 14px 16px;
    background: #f3f4f6;
    border-radius: 8px;
    overflow-x: auto;
    font-size: 0.9rem;
    line-height: 1.55;
    white-space: pre;
    tab-size: 4;
  }

  .portal-root .article .md-body pre code,
  .portal-root .article .html-body pre code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: inherit;
    background: transparent;
    padding: 0;
    border-radius: 0;
    color: inherit;
  }

  .portal-root .article .md-body :not(pre) > code,
  .portal-root .article .html-body :not(pre) > code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.9em;
    background: #f3f4f6;
    padding: 0.15em 0.45em;
    border-radius: 4px;
    color: #b45309;
  }

  .portal-root .article .md-body strong,
  .portal-root .article .html-body strong {
    font-weight: 600;
    color: #111827;
  }

  .portal-root .article .md-body em,
  .portal-root .article .html-body em {
    font-style: italic;
  }
</style>
