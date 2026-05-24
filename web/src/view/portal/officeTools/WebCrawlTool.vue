<template>
  <div class="tool-panel">
    <CrawlerDisclaimerAlert />
    <el-alert
      title="输入商品列表页或详情页网址，解析 JSON-LD / 常见 HTML 结构，导出 Excel（名称、价格、图片链接、商家、来源网址）。纯 JS 渲染站点可能无法识别。"
      type="info"
      show-icon
      :closable="false"
      class="mb-3"
    />
    <el-input v-model="url" placeholder="https://shop.example.com/products" clearable>
      <template #prepend>网址</template>
    </el-input>
    <el-button type="primary" class="mt-3" :loading="loading" :disabled="!url.trim()" @click="run">
      爬取并下载 Excel
    </el-button>
    <p class="hint mt-2">单次最多约 200 条；请确保目标站允许抓取且为静态或可解析结构。</p>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { crawlWebProductsExcel } from '@/api/publicPortalOffice'
  import CrawlerDisclaimerAlert from '@/view/portal/components/CrawlerDisclaimerAlert.vue'

  const url = ref('')
  const loading = ref(false)

  const run = async () => {
    loading.value = true
    try {
      const res = await crawlWebProductsExcel({ url: url.value.trim() })
      const blob = res?.data instanceof Blob ? res.data : res
      if (blob?.type?.includes('json')) {
        const t = await blob.text()
        try {
          ElMessage.error(JSON.parse(t).msg || '爬取失败')
        } catch {
          ElMessage.error('爬取失败')
        }
        return
      }
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = 'product-crawl.xlsx'
      a.click()
      URL.revokeObjectURL(a.href)
      ElMessage.success('Excel 已开始下载')
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '爬取失败')
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
</style>
