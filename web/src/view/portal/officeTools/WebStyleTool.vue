<template>
  <div class="tool-panel">
    <el-alert
      title="输入网址后，自动爬取同域下的前端页面（HTML）及关联的 CSS/JS/图片等静态资源，打包为 ZIP 下载。不做仿站生成，仅原样保存可抓取的页面。"
      type="info"
      show-icon
      :closable="false"
      class="mb-3"
    />
    <el-input v-model="url" placeholder="https://example.com" clearable>
      <template #prepend>起始网址</template>
    </el-input>
    <el-row :gutter="12" class="mt-2">
      <el-col :span="12">
        <span class="label">最大页面数</span>
        <el-input-number v-model="maxPages" :min="1" :max="200" class="w-full" />
      </el-col>
      <el-col :span="12">
        <span class="label">爬取深度</span>
        <el-input-number v-model="maxDepth" :min="1" :max="10" class="w-full" />
      </el-col>
    </el-row>
    <el-button type="primary" class="mt-3" :loading="loading" :disabled="!url.trim()" @click="run">
      下载网站前端页面 ZIP
    </el-button>
    <p class="hint mt-2">
      仅同域链接；默认最多 50 页、深度 3；纯 Vue/React 单页应用可能只能下到壳页面。服务器临时数据约 24 小时后清理。
    </p>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { downloadWebsitePagesZip } from '@/api/publicPortalOffice'

  const url = ref('')
  const maxPages = ref(50)
  const maxDepth = ref(3)
  const loading = ref(false)

  const run = async () => {
    loading.value = true
    try {
      const res = await downloadWebsitePagesZip({
        url: url.value.trim(),
        maxPages: maxPages.value,
        maxDepth: maxDepth.value
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
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = 'website-pages.zip'
      a.click()
      URL.revokeObjectURL(a.href)
      ElMessage.success('ZIP 已开始下载')
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '下载失败')
    } finally {
      loading.value = false
    }
  }
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .mb-3 { margin-bottom: 12px; }
  .mt-2 { margin-top: 8px; }
  .mt-3 { margin-top: 12px; }
  .hint { font-size: 12px; color: #909399; }
  .label { font-size: 12px; color: #606266; display: block; margin-bottom: 4px; }
  .w-full { width: 100%; }
</style>
