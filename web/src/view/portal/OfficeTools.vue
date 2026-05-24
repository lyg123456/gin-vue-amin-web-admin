<template>

  <div class="office-tools-page">

    <section class="portal-panel tools-hub">

      <div class="page-title-bar">

        <h1 class="page-h1">{{ activeMeta.title }}</h1>

        <p class="page-desc">本地处理为主 · 临时数据约 24 小时自动清理</p>

      </div>



      <nav class="mega-nav" aria-label="工具导航">

        <button

          v-for="c in categories"

          :key="c.id"

          type="button"

          class="mega-item"

          :class="{ active: activeCat === c.id && !isQuickActive }"

          @click="selectCategory(c.id)"

        >

          {{ c.name }}

        </button>

        <span class="mega-divider" aria-hidden="true" />

        <button

          v-for="t in quickTools"

          :key="t.id"

          type="button"

          class="mega-item mega-item--quick"

          :class="{ active: activeId === t.id }"

          @click="selectTool(t.id)"

        >

          {{ t.title }}

        </button>

      </nav>



      <div v-if="!isQuickActive" class="tool-links">

        <button

          v-for="t in currentTools"

          :key="t.id"

          type="button"

          class="tool-link"

          :class="{ active: activeId === t.id }"

          @click="selectTool(t.id)"

        >

          {{ t.title }}

          <span v-if="t.hot" class="hot-tag">热</span>

        </button>

      </div>

    </section>



    <section v-if="activeId" class="portal-panel tool-workspace">

      <JsonFormatterTool v-if="activeId === 'json'" />

      <CodecTool v-else-if="activeId === 'codec'" />

      <TimeConverterTool v-else-if="activeId === 'time'" />

      <NamingTool v-else-if="activeId === 'naming'" />

      <CodeFormatTool v-else-if="activeId === 'codefmt'" />

      <TextDiffTool v-else-if="activeId === 'diff'" />

      <TextCleanTool v-else-if="activeId === 'textclean'" />

      <MarkdownTool v-else-if="activeId === 'markdown'" />

      <CompressToolsTool v-else-if="activeId === 'compress'" />

      <TempEmailTool v-else-if="activeId === 'email'" />

      <FileConvertTool v-else-if="activeId === 'file'" />

      <MediaToolsTool v-else-if="activeId === 'media'" />

      <TweetLayoutTool v-else-if="activeId === 'tweet'" />

      <WebStyleTool v-else-if="activeId === 'webstyle'" />

      <WebCrawlTool v-else-if="activeId === 'webcrawl'" />
      <NetworkSpeedTool v-else-if="activeId === 'speed'" />
      <WatermarkTool v-else-if="activeId === 'watermark'" />
      <DouyinIndustryTool v-else-if="activeId === 'douyin'" />
      <WechatVideoCrawlTool v-else-if="activeId === 'wechat'" />
      <XhsVideoCrawlTool v-else-if="activeId === 'xhs'" />

    </section>

  </div>

</template>



<script setup>

  import { computed, onMounted, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'

  import TimeConverterTool from '@/view/portal/officeTools/TimeConverterTool.vue'

  import TempEmailTool from '@/view/portal/officeTools/TempEmailTool.vue'

  import JsonFormatterTool from '@/view/portal/officeTools/JsonFormatterTool.vue'

  import CodecTool from '@/view/portal/officeTools/CodecTool.vue'

  import NamingTool from '@/view/portal/officeTools/NamingTool.vue'

  import CodeFormatTool from '@/view/portal/officeTools/CodeFormatTool.vue'

  import TextDiffTool from '@/view/portal/officeTools/TextDiffTool.vue'

  import TextCleanTool from '@/view/portal/officeTools/TextCleanTool.vue'

  import MarkdownTool from '@/view/portal/officeTools/MarkdownTool.vue'

  import FileConvertTool from '@/view/portal/officeTools/FileConvertTool.vue'

  import MediaToolsTool from '@/view/portal/officeTools/MediaToolsTool.vue'

  import CompressToolsTool from '@/view/portal/officeTools/CompressToolsTool.vue'

  import TweetLayoutTool from '@/view/portal/officeTools/TweetLayoutTool.vue'

  import WebStyleTool from '@/view/portal/officeTools/WebStyleTool.vue'

  import WebCrawlTool from '@/view/portal/officeTools/WebCrawlTool.vue'
  import NetworkSpeedTool from '@/view/portal/officeTools/NetworkSpeedTool.vue'
  import WatermarkTool from '@/view/portal/officeTools/WatermarkTool.vue'
  import DouyinIndustryTool from '@/view/portal/officeTools/DouyinIndustryTool.vue'
  import WechatVideoCrawlTool from '@/view/portal/officeTools/WechatVideoCrawlTool.vue'
  import XhsVideoCrawlTool from '@/view/portal/officeTools/XhsVideoCrawlTool.vue'
  import { VIDEO_CRAWL_TOOL_IDS } from '@/constants/videoCrawlNav'

  defineOptions({ name: 'WebOfficeTools' })

  const route = useRoute()



  const quickTools = [
    { id: 'email', title: '临时邮箱' },
    { id: 'speed', title: '网络测速' },
    { id: 'watermark', title: '去除水印' },
    { id: 'webstyle', title: '网站页面下载' },
    { id: 'webcrawl', title: '商品爬取 Excel' },
    { id: 'tweet', title: '推文排版' },
    { id: 'douyin', title: '抖音垂类抓取' },
    { id: 'xhs', title: '小红书爆款抓取' }
  ]



  const quickToolIds = new Set([...quickTools.map((t) => t.id), ...VIDEO_CRAWL_TOOL_IDS])



  const categories = [

    {

      id: 'json',

      name: 'JSON工具',

      tools: [

        { id: 'json', title: 'JSON在线解析', hot: true },

        { id: 'codec', title: 'Base64 / URL / MD5', hot: true }

      ]

    },

    {

      id: 'format',

      name: '压缩/格式化',

      tools: [

        { id: 'codefmt', title: 'JS / CSS 格式化', hot: true },

        { id: 'markdown', title: 'Markdown 预览', hot: true }

      ]

    },

    {

      id: 'common',

      name: '常用工具',

      tools: [

        { id: 'time', title: 'Unix 时间戳', hot: true },

        { id: 'naming', title: 'UUID / 密码 / 命名', hot: true },

        { id: 'diff', title: '文本对比', hot: true },

        { id: 'textclean', title: '文本整理去重', hot: true }

      ]

    },

    {

      id: 'image',

      name: '图片/文件',

      tools: [

        { id: 'compress', title: '图片 / Excel 压缩', hot: true },

        { id: 'file', title: '文件 PDF 转换', hot: true },

        { id: 'media', title: '音视频图像' }

      ]

    }

  ]



  const titleMap = {

    json: 'JSON在线解析、JSON格式化',

    codec: '编码 / 加密 / 解密',

    time: 'Unix 时间戳转换',

    naming: 'UUID · 密码 · 命名转换',

    codefmt: 'JS / CSS 格式化压缩',

    diff: '文本对比',

    textclean: '文本整理',

    markdown: 'Markdown 编辑器',

    compress: '文件压缩',

    email: '临时邮箱',

    file: 'Office 文件转换',

    media: '音视频图像处理',

    tweet: '推文排版',

    webstyle: '网站页面下载',

    webcrawl: '商品爬取 Excel',
    douyin: '抖音垂类抓取',
    wechat: '微信视频抓取',
    xhs: '小红书视频抓取',
    speed: '网络测速',
    watermark: '去除水印'

  }



  const activeCat = ref('json')

  const activeId = ref('json')



  const currentTools = computed(() => categories.find((c) => c.id === activeCat.value)?.tools || [])



  const isQuickActive = computed(() => quickToolIds.has(activeId.value))



  const activeMeta = computed(() => ({ title: titleMap[activeId.value] || '在线工具' }))



  const selectCategory = (catId) => {

    activeCat.value = catId

    const list = categories.find((c) => c.id === catId)?.tools || []

    if (list.length) activeId.value = list[0].id

  }



  const selectTool = (id) => {

    activeId.value = id

  }

  const applyToolFromRoute = () => {
    const tool = String(route.query.tool || '')
    if (VIDEO_CRAWL_TOOL_IDS.has(tool)) {
      activeId.value = tool
    }
  }

  onMounted(applyToolFromRoute)
  watch(() => route.query.tool, applyToolFromRoute)

</script>



<style lang="scss" scoped>

  .office-tools-page {

    display: flex;

    flex-direction: column;

    gap: 10px;

    padding-bottom: 16px;

  }



  .portal-panel {

    background: var(--portal-panel-bg, #fff);

    border-radius: var(--portal-radius, 4px);

    padding: 14px 16px;

    box-sizing: border-box;

    border: 1px solid #e0e0e0;

  }



  .tools-hub {

    padding: 0;

    overflow: hidden;

  }



  .page-title-bar {

    padding: 14px 16px 10px;

    border-bottom: 1px solid #eee;

    background: #fafafa;

  }



  .page-h1 {

    margin: 0;

    font-size: var(--portal-font-title, 18px);

    font-weight: 700;

    color: #333;

    line-height: 1.4;

  }



  .page-desc {

    margin: 6px 0 0;

    font-size: 13px;

    color: #888;

  }



  /* 分类 + 快捷工具同一行，自动换行 */

  .mega-nav {

    display: flex;

    flex-wrap: wrap;

    align-items: center;

    gap: 0;

    background: #fff;

    border-bottom: 1px solid #e0e0e0;

    padding: 4px 8px 0;

  }



  .mega-item {

    border: none;

    background: transparent;

    padding: 12px 14px;

    font-size: 15px;

    color: #333;

    cursor: pointer;

    border-bottom: 2px solid transparent;

    margin-bottom: -1px;

    white-space: nowrap;

    transition: color 0.15s, border-color 0.15s;



    &:hover {

      color: var(--portal-brand, #1a73e8);

    }



    &.active {

      color: var(--portal-brand, #1a73e8);

      font-weight: 600;

      border-bottom-color: var(--portal-brand, #1a73e8);

    }

  }



  .mega-item--quick {

    font-size: 14px;

  }



  .mega-divider {

    width: 1px;

    height: 20px;

    background: #ddd;

    margin: 0 6px 8px;

    flex-shrink: 0;

  }



  .tool-links {

    display: flex;

    flex-wrap: wrap;

    gap: 10px 12px;

    padding: 14px 16px;

    background: #fff;

  }



  .tool-link {

    display: inline-flex;

    align-items: center;

    gap: 6px;

    padding: 8px 14px;

    min-height: 36px;

    border: 1px solid #d9d9d9;

    border-radius: 3px;

    background: #fff;

    cursor: pointer;

    font-size: 14px;

    color: #333;

    transition: all 0.15s;



    &:hover {

      border-color: var(--portal-brand, #1a73e8);

      color: var(--portal-brand, #1a73e8);

    }



    &.active {

      border-color: var(--portal-brand, #1a73e8);

      background: #e8f4ff;

      color: var(--portal-brand-dark, #1557b0);

      font-weight: 600;

    }

  }



  .hot-tag {

    font-size: 12px;

    padding: 0 5px;

    line-height: 18px;

    border-radius: 2px;

    background: #ff6600;

    color: #fff;

    font-weight: normal;

  }



  .tool-workspace {

    padding: 16px;

    min-height: 200px;

  }



  @media (max-width: 640px) {

    .mega-nav {

      padding: 4px 4px 0;

    }



    .mega-item {

      font-size: 14px;

      padding: 10px 10px;

    }



    .mega-divider {

      display: none;

    }



    .tool-link {

      font-size: 13px;

    }

  }

</style>

