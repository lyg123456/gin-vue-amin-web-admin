<template>
  <div class="tool-panel">
    <CrawlerDisclaimerAlert />
    <el-alert
      type="info"
      show-icon
      :closable="false"
      class="mb-3"
    >
      <template #title>
        登录
        <a href="https://channels.weixin.qq.com/" target="_blank" rel="noopener">channels.weixin.qq.com</a>
        仅抓取其他创作者作品（播放/点赞）。微信无抖音式全站热榜，需通过搜索发现他人视频。
        推荐：本机运行
        <a href="https://github.com/nobiyou/wx_channel" target="_blank" rel="noopener">wx_channel</a>
        代理，浏览器打开视频号页面后填写代理地址；或粘贴助手 Cookie 作备用。
      </template>
    </el-alert>

    <el-form label-width="108px" class="form-block">
      <el-form-item label="wx_channel">
        <el-input
          v-model="wxChannelBase"
          placeholder="http://127.0.0.1:2026（推荐，用于搜索他人视频号）"
          clearable
        />
        <div class="hint wx-hint">
          <p>1. 从 <a href="https://github.com/nobiyou/wx_channel/releases" target="_blank" rel="noopener">GitHub Releases</a> 下载并双击运行 <code>wx_channel.exe</code>（窗口勿关）</p>
          <p>2. 默认 API 地址 <code>http://127.0.0.1:2026</code>（启动日志里若写了别的端口请改成「代理端口+1」）</p>
          <p>3. 用 Chrome 打开视频号页面；后台须与 wx_channel 在同一台电脑（远程服务器上的 127.0.0.1 连不到你本机）</p>
          <p>4. 浏览器访问 <code>http://127.0.0.1:2026/api/channels/status</code> 有 JSON 后再点「检测」</p>
        </div>
      </el-form-item>

      <el-form-item label="Cookie">
        <el-input
          v-model="cookie"
          type="textarea"
          :rows="4"
          placeholder="从视频号助手页面复制 Cookie..."
          @blur="saveCookieLocal"
        />
        <div class="cookie-actions">
          <el-button size="small" :loading="verifying" @click="verifyCookie">检测 Cookie</el-button>
          <el-button size="small" text type="danger" @click="clearCookie">清除本地缓存</el-button>
          <span v-if="verifyInfo" class="verify-text" :class="{ ok: verifyInfo.ok }">
            {{ verifyInfo.message }}
            <template v-if="verifyInfo.nickname">（{{ verifyInfo.nickname }}）</template>
          </span>
        </div>
      </el-form-item>

      <el-form-item label="官方垂类">
        <div class="preset-row">
          <el-button size="small" @click="applyPreset('hot6')">常用 6 类</el-button>
          <el-button size="small" @click="applyPreset('all')">全选</el-button>
          <el-button size="small" @click="selectedIds = []">清空</el-button>
        </div>
        <el-select
          v-model="selectedIds"
          multiple
          filterable
          collapse-tags
          collapse-tags-tooltip
          placeholder="选择垂类"
          style="width: 100%; margin-top: 8px"
        >
          <el-option
            v-for="c in categories"
            :key="c.id"
            :label="`${c.name}（${c.id}）`"
            :value="c.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="热度指标">
        <el-radio-group v-model="metric">
          <el-radio value="auto">自动（有播放量用播放，否则用点赞）</el-radio>
          <el-radio value="play">仅播放量</el-radio>
          <el-radio value="digg">仅点赞数</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="热度下限">
        <el-input-number v-model="minPlayCount" :min="1000" :max="50000000" :step="1000" />
        <span class="hint inline">推荐 20000；部分作品仅展示点赞时可改「自动」</span>
      </el-form-item>

      <el-form-item label="每类条数">
        <el-input-number v-model="limitPerCat" :min="1" :max="20" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="loading" :disabled="!canRun" @click="run">
          开始抓取
        </el-button>
        <el-button :disabled="!results.length" @click="exportJson">导出 JSON</el-button>
      </el-form-item>
    </el-form>

    <el-alert v-if="globalNote" :title="globalNote" type="warning" show-icon class="mb-3" />

    <div v-if="results.length" class="results">
      <div v-for="block in results" :key="block.categoryId" class="cat-block">
        <h3 class="cat-title">
          {{ block.categoryName }}
          <span class="cat-id">#{{ block.categoryId }}</span>
          <el-tag v-if="block.source" size="small" type="info">{{ sourceLabel(block.source) }}</el-tag>
          <el-tag v-if="block.error" type="danger" size="small">{{ block.error }}</el-tag>
          <el-tag v-else type="success" size="small">{{ block.videos?.length || 0 }} 条</el-tag>
        </h3>
        <el-table v-if="block.videos?.length" :data="block.videos" stripe size="small">
          <el-table-column label="封面" width="72">
            <template #default="{ row }">
              <img v-if="row.coverUrl" :src="row.coverUrl" class="cover-thumb" alt="" />
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
          <el-table-column prop="authorName" label="作者" width="100" show-overflow-tooltip />
          <el-table-column label="热度" width="120">
            <template #default="{ row }">
              <span>{{ formatNum(row.effectiveStat) }}</span>
              <el-tag size="small" class="ml-1">{{ row.statSource === 'digg' ? '赞' : '播' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="playCount" label="播放" width="88">
            <template #default="{ row }">{{ formatNum(row.playCount) }}</template>
          </el-table-column>
          <el-table-column prop="diggCount" label="点赞" width="88">
            <template #default="{ row }">{{ formatNum(row.diggCount) }}</template>
          </el-table-column>
          <el-table-column label="链接" width="64">
            <template #default="{ row }">
              <a v-if="row.videoUrl" :href="row.videoUrl" target="_blank" rel="noopener">打开</a>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    crawlWechatIndustryVideos,
    getWechatOfficialCategories,
    verifyWechatCookie
  } from '@/api/publicPortalOffice'
  import CrawlerDisclaimerAlert from '@/view/portal/components/CrawlerDisclaimerAlert.vue'

  defineOptions({ name: 'WechatVideoCrawlTool' })

  const STORAGE_KEY = 'dnf1688_wechat_cookie'
  const WX_CHANNEL_KEY = 'dnf1688_wx_channel_base'

  const cookie = ref('')
  const wxChannelBase = ref('http://127.0.0.1:2026')
  const categories = ref([])
  const selectedIds = ref([2, 3, 8, 19, 1, 7])
  const minPlayCount = ref(20000)
  const limitPerCat = ref(5)
  const metric = ref('auto')
  const loading = ref(false)
  const verifying = ref(false)
  const verifyInfo = ref(null)
  const results = ref([])
  const globalNote = ref('')

  const canRun = computed(
    () =>
      (cookie.value.trim() || wxChannelBase.value.trim()) &&
      selectedIds.value.length > 0 &&
      !loading.value
  )

  const formatNum = (n) => {
    if (n == null || n === 0) return '-'
    if (n >= 10000) return (n / 10000).toFixed(1) + '万'
    return String(n)
  }

  const sourceLabel = (s) => {
    if (s === 'others') return '他人作品'
    if (s === 'video_search') return '关键词搜视频'
    if (s === 'account_feed') return '他人账号作品'
    return s || '他人'
  }

  const saveCookieLocal = () => {
    const v = cookie.value.trim()
    if (v) localStorage.setItem(STORAGE_KEY, v)
  }

  const clearCookie = () => {
    cookie.value = ''
    localStorage.removeItem(STORAGE_KEY)
    verifyInfo.value = null
    ElMessage.success('已清除')
  }

  const applyPreset = (name) => {
    if (name === 'hot6') {
      selectedIds.value = [2, 3, 8, 19, 1, 7]
    } else if (name === 'all') {
      selectedIds.value = categories.value.map((c) => c.id)
    }
  }

  const verifyCookie = async () => {
    if (!cookie.value.trim()) {
      ElMessage.warning('请先粘贴 Cookie')
      return
    }
    verifying.value = true
    verifyInfo.value = null
    try {
      const res = await verifyWechatCookie({
        cookie: cookie.value.trim(),
        wxChannelBase: wxChannelBase.value.trim()
      })
      verifyInfo.value = res.data || {}
      saveCookieLocal()
      if (verifyInfo.value.ok) {
        ElMessage.success(verifyInfo.value.message || 'Cookie 有效')
      } else {
        ElMessage.warning(verifyInfo.value.message || 'Cookie 无效')
      }
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '检测失败')
    } finally {
      verifying.value = false
    }
  }

  onMounted(async () => {
    const cached = localStorage.getItem(STORAGE_KEY)
    if (cached) cookie.value = cached
    const wxCached = localStorage.getItem(WX_CHANNEL_KEY)
    if (wxCached) wxChannelBase.value = wxCached
    try {
      const res = await getWechatOfficialCategories()
      categories.value = res.data || []
    } catch (e) {
      ElMessage.error(e?.message || '加载垂类失败')
    }
  })

  const run = async () => {
    loading.value = true
    results.value = []
    globalNote.value = ''
    saveCookieLocal()
    try {
      if (wxChannelBase.value.trim()) {
        localStorage.setItem(WX_CHANNEL_KEY, wxChannelBase.value.trim())
      }
      const res = await crawlWechatIndustryVideos({
        cookie: cookie.value.trim(),
        wxChannelBase: wxChannelBase.value.trim(),
        categoryIds: selectedIds.value,
        minPlayCount: minPlayCount.value,
        limitPerCat: limitPerCat.value,
        metric: metric.value
      })
      const payload = res?.data
      const list = Array.isArray(payload)
        ? payload
        : Array.isArray(payload?.list)
          ? payload.list
          : []
      results.value = list
      globalNote.value = (payload && !Array.isArray(payload) ? payload.note : '') || ''
      const ok = list.filter((r) => r.videos?.length).length
      if (ok === 0) {
        ElMessage.warning('未抓到他人热门视频：确认 wx_channel 已连接并打开视频号页，或降低热度下限/换垂类')
      } else {
        ElMessage.success(`完成 ${ok}/${list.length} 个垂类`)
      }
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '抓取失败')
    } finally {
      loading.value = false
    }
  }

  const exportJson = () => {
    const blob = new Blob(
      [JSON.stringify({ note: globalNote.value, results: results.value }, null, 2)],
      { type: 'application/json' }
    )
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `wechat-video-crawl-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(a.href)
  }
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .mb-3 { margin-bottom: 12px; }
  .hint { font-size: 12px; color: #909399; }
  .hint.inline { margin-left: 8px; }
  .cookie-actions { margin-top: 8px; display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  .verify-text { font-size: 12px; color: #909399; max-width: 100%; }
  .verify-text.ok { color: #67c23a; }
  .preset-row { display: flex; gap: 8px; flex-wrap: wrap; }
  .results { margin-top: 16px; }
  .cat-block { margin-bottom: 20px; }
  .cat-title {
    font-size: 15px;
    margin: 0 0 8px;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .cat-id { color: #909399; font-weight: normal; font-size: 12px; }
  .cover-thumb { width: 48px; height: 48px; object-fit: cover; border-radius: 4px; }
  .ml-1 { margin-left: 4px; }
  .wx-hint { line-height: 1.6; }
  .wx-hint p { margin: 4px 0; }
  .wx-hint code { font-size: 12px; background: #f4f4f5; padding: 0 4px; border-radius: 3px; }
</style>
