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
        登录 douyin.com → F12 → 网络 → 任选请求 → 复制完整 Cookie。需含
        <code>sessionid</code>、<code>ttwid</code>。每垂类 5 条，热度默认 ≥2 万。
      </template>
    </el-alert>

    <el-form label-width="108px" class="form-block">
      <el-form-item label="Cookie" required>
        <el-input
          v-model="cookie"
          type="textarea"
          :rows="4"
          placeholder="sessionid=...; ttwid=...; passport_csrf_token=..."
          @blur="saveCookieLocal"
        />
        <div class="cookie-actions">
          <el-button size="small" :loading="verifying" @click="verifyCookie">检测 Cookie</el-button>
          <el-button size="small" text type="danger" @click="clearCookie">清除本地缓存</el-button>
          <span v-if="verifyInfo" class="verify-text" :class="{ ok: verifyInfo.ok, warn: verifyInfo.searchNeedVerify }">
            {{ verifyInfo.message }}
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
        <span class="hint inline">推荐 20000；平台常隐藏播放量时可改「自动」或改用点赞</span>
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
          <el-table-column label="视频页" width="64">
            <template #default="{ row }">
              <a v-if="row.pageUrl || row.videoUrl" :href="row.pageUrl || row.videoUrl" target="_blank" rel="noopener">打开</a>
            </template>
          </el-table-column>
          <el-table-column label="下载" width="148">
            <template #default="{ row }">
              <template v-if="row.downloadUrl">
                <el-button
                  size="small"
                  type="primary"
                  :loading="downloadingId === row.awemeId"
                  @click="downloadVideo(row)"
                >
                  下载
                </el-button>
                <el-button size="small" link @click="copyText(row.downloadUrl)">复制链接</el-button>
              </template>
              <span v-else class="hint">未解析到直链</span>
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
    crawlDouyinIndustryVideos,
    getDouyinOfficialCategories,
    proxyOfficeMediaDownload,
    verifyDouyinCookie
  } from '@/api/publicPortalOffice'
  import CrawlerDisclaimerAlert from '@/view/portal/components/CrawlerDisclaimerAlert.vue'

  const STORAGE_KEY = 'dnf1688_douyin_cookie'

  const cookie = ref('')
  const categories = ref([])
  const selectedIds = ref([603, 628, 604, 613, 619, 629])
  const minPlayCount = ref(20000)
  const limitPerCat = ref(5)
  const metric = ref('auto')
  const loading = ref(false)
  const verifying = ref(false)
  const verifyInfo = ref(null)
  const results = ref([])
  const globalNote = ref('')
  const downloadingId = ref('')

  const canRun = computed(
    () => cookie.value.trim() && selectedIds.value.length > 0 && !loading.value
  )

  const downloadVideo = async (row) => {
    if (!cookie.value.trim()) {
      ElMessage.warning('请先填写 Cookie 再下载')
      return
    }
    if (!row.downloadUrl) return
    downloadingId.value = row.awemeId
    try {
      const res = await proxyOfficeMediaDownload({
        url: row.downloadUrl,
        cookie: cookie.value.trim(),
        platform: 'douyin',
        title: row.title || row.awemeId
      })
      const blob = res?.data instanceof Blob ? res.data : res
      if (blob?.type?.includes('json')) {
        const t = await blob.text()
        try {
          ElMessage.error(JSON.parse(t).msg || '下载失败')
        } catch {
          ElMessage.error('下载失败')
        }
        return
      }
      const safe = (row.title || row.awemeId || 'douyin-video').replace(/[<>:"/\\|?*]/g, '_').slice(0, 60)
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = safe + '.mp4'
      a.click()
      URL.revokeObjectURL(a.href)
      ElMessage.success('视频已开始下载')
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '下载失败，请重新抓取后再试')
    } finally {
      downloadingId.value = ''
    }
  }

  const copyText = async (text) => {
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制下载链接')
    } catch {
      ElMessage.warning('复制失败，请手动选中链接复制')
    }
  }

  const formatNum = (n) => {
    if (n == null || n === 0) return '-'
    if (n >= 10000) return (n / 10000).toFixed(1) + '万'
    return String(n)
  }

  const sourceLabel = (s) => (s === 'search' ? '垂类搜索' : '推荐流匹配')

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
      selectedIds.value = [603, 628, 604, 613, 619, 629]
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
      const res = await verifyDouyinCookie({ cookie: cookie.value.trim() })
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
    try {
      const res = await getDouyinOfficialCategories()
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
      const res = await crawlDouyinIndustryVideos({
        cookie: cookie.value.trim(),
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
        ElMessage.warning('未抓到数据：先点「检测 Cookie」，或在浏览器完成搜索人机验证后重试')
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
    a.download = `douyin-crawl-${Date.now()}.json`
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
  .verify-text.warn { color: #e6a23c; }
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
  code { font-size: 12px; background: #f4f4f5; padding: 0 4px; border-radius: 3px; }
</style>
