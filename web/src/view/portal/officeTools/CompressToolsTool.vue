<template>
  <div class="tool-panel">
    <el-tabs v-model="tab">
      <el-tab-pane label="图片压缩" name="image">
        <p class="hint">服务端压缩：可调质量与最大宽度，下载体积更小</p>
        <el-upload drag :auto-upload="false" :limit="1" accept="image/*" @change="(f) => (imgFile = f?.raw)">
          <div>拖拽或点击选择图片</div>
        </el-upload>
        <el-row :gutter="12" class="mt-3">
          <el-col :span="12">
            <span class="label">JPEG 质量 (1-100)</span>
            <el-slider v-model="quality" :min="30" :max="100" />
          </el-col>
          <el-col :span="12">
            <span class="label">最大宽度 (px)</span>
            <el-input-number v-model="maxWidth" :min="320" :max="4096" class="w-full" />
          </el-col>
        </el-row>
        <el-button type="primary" class="mt-3" :loading="loading" :disabled="!imgFile" @click="compressImg">压缩并下载</el-button>
      </el-tab-pane>
      <el-tab-pane label="Excel 压缩" name="excel">
        <p class="hint">重建 xlsx 仅保留单元格数据并 ZIP 压缩，适合去样式减负</p>
        <el-upload :auto-upload="false" :limit="1" accept=".xlsx,.xls" @change="(f) => (excelFile = f?.raw)">
          <el-button type="primary">选择 Excel</el-button>
        </el-upload>
        <el-button type="primary" class="mt-3" :loading="loading" :disabled="!excelFile" @click="compressXls">压缩并下载</el-button>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { compressOfficeExcel, compressOfficeImage } from '@/api/publicPortalOffice'

  const tab = ref('image')
  const loading = ref(false)
  const imgFile = ref(null)
  const excelFile = ref(null)
  const quality = ref(80)
  const maxWidth = ref(1920)

  const downloadBlob = (res, name) => {
    const blob = res?.data instanceof Blob ? res.data : res
    if (blob?.type?.includes('json')) {
      blob.text().then((t) => {
        try {
          ElMessage.error(JSON.parse(t).msg || '失败')
        } catch {
          ElMessage.error('失败')
        }
      })
      return
    }
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = name
    a.click()
    URL.revokeObjectURL(a.href)
    const o = res?.headers?.['x-original-size']
    const n = res?.headers?.['x-new-size']
    if (o && n) ElMessage.success(`压缩完成：${n} / ${o} 字节`)
    else ElMessage.success('已开始下载')
  }

  const compressImg = async () => {
    loading.value = true
    try {
      const fd = new FormData()
      fd.append('file', imgFile.value)
      fd.append('quality', String(quality.value))
      fd.append('maxWidth', String(maxWidth.value))
      const res = await compressOfficeImage(fd)
      downloadBlob(res, 'compressed.jpg')
    } catch (e) {
      ElMessage.error(e?.message || '压缩失败')
    } finally {
      loading.value = false
    }
  }

  const compressXls = async () => {
    loading.value = true
    try {
      const fd = new FormData()
      fd.append('file', excelFile.value)
      const res = await compressOfficeExcel(fd)
      downloadBlob(res, 'compressed.xlsx')
    } catch (e) {
      ElMessage.error(e?.message || '压缩失败')
    } finally {
      loading.value = false
    }
  }
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .hint { font-size: 13px; color: #909399; margin: 0 0 12px; }
  .mt-3 { margin-top: 12px; }
  .label { font-size: 12px; color: #606266; }
  .w-full { width: 100%; }
</style>
