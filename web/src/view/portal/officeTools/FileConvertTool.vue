<template>
  <div class="tool-panel">
    <el-alert
      v-if="caps"
      :title="capHint"
      :type="caps.libreOffice ? 'success' : 'warning'"
      show-icon
      :closable="false"
      class="mb-3"
    />
    <el-tabs v-model="tab">
      <el-tab-pane label="图片转换" name="image">
        <p class="hint">上传图片，由服务端转换为 PNG / JPEG；WebP 可在浏览器本地转换（不上传）。</p>
        <el-upload drag :auto-upload="false" :show-file-list="false" accept="image/*" @change="onImagePick">
          <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
          <div>选择图片（服务端转换）</div>
        </el-upload>
        <el-radio-group v-model="imageTarget" class="mt-3">
          <el-radio-button label="png">PNG</el-radio-button>
          <el-radio-button label="jpeg">JPEG</el-radio-button>
          <el-radio-button label="webp-local">WebP（本地）</el-radio-button>
        </el-radio-group>
        <el-button type="primary" class="mt-3" :loading="converting" :disabled="!imageFile" @click="convertImage">
          转换并下载
        </el-button>
      </el-tab-pane>

      <el-tab-pane label="Office → PDF" name="office">
        <p class="hint">
          Gin 后端已实现转换。高保真（版式/图片/PPT）需服务器安装 LibreOffice；
          未安装时 <strong>.docx / .xlsx</strong> 可用 Go 简易转 PDF（文字/表格）。
        </p>
        <el-upload :auto-upload="false" :limit="1" accept=".doc,.docx,.xls,.xlsx,.ppt,.pptx,.odt,.ods,.odp,.rtf" @change="onOfficePick">
          <el-button type="primary">选择 Office 文件</el-button>
        </el-upload>
        <p v-if="officeFile" class="file-name">{{ officeFile.name }}</p>
        <el-button type="primary" class="mt-3" :loading="converting" :disabled="!officeFile || !canOfficePdf" @click="convertOffice">
          转为 PDF 并下载
        </el-button>
      </el-tab-pane>

      <el-tab-pane label="文本 → PDF" name="textpdf">
        <el-input v-model="textContent" type="textarea" :rows="10" placeholder="输入文本，由服务端生成 PDF" />
        <el-button type="primary" class="mt-3" :loading="converting" @click="convertTextPdf">生成 PDF</el-button>
      </el-tab-pane>

      <el-tab-pane label="图片 → PDF" name="imgpdf">
        <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" @change="onImgPdfPick">
          <el-button>选择图片</el-button>
        </el-upload>
        <el-button type="primary" class="mt-3" :loading="converting" :disabled="!imgPdfFile" @click="convertImgPdf">导出 PDF</el-button>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { UploadFilled } from '@element-plus/icons-vue'
  import { convertOfficeFile, getOfficeConvertCapabilities } from '@/api/publicPortalOffice'

  const tab = ref('image')
  const caps = ref(null)
  const converting = ref(false)
  const imageFile = ref(null)
  const imageTarget = ref('png')
  const officeFile = ref(null)
  const imgPdfFile = ref(null)
  const textContent = ref('在此输入要导出为 PDF 的文本…')

  const canOfficePdf = computed(() => {
    if (!caps.value) return true
    return caps.value.officePdfAvailable !== false
  })

  const capHint = computed(() => {
    if (!caps.value) return ''
    const parts = []
    if (caps.value.libreOffice) {
      parts.push(`LibreOffice 高保真：${caps.value.libreOfficePath || '已就绪'}`)
    } else if (caps.value.libreOfficeHint) {
      parts.push(caps.value.libreOfficeHint)
    }
    if (caps.value.goOfficeFallback) {
      parts.push(`Go 简易：${(caps.value.goFallbackExts || ['.docx', '.xlsx']).join(' ')}`)
    }
    parts.push(`单文件 ≤ ${caps.value.maxUploadMB || 20}MB`)
    return parts.join('；')
  })

  const downloadBlob = (blob, name) => {
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = name || 'download'
    a.click()
    URL.revokeObjectURL(a.href)
  }

  const postConvert = async (formData, fallbackName) => {
    converting.value = true
    try {
      const res = await convertOfficeFile(formData)
      const blob = res?.data instanceof Blob ? res.data : res
      if (blob?.type?.includes('application/json')) {
        const text = await blob.text()
        const j = JSON.parse(text)
        throw new Error(j.msg || '转换失败')
      }
      let name = fallbackName
      downloadBlob(blob, name)
      ElMessage.success('已开始下载')
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '转换失败')
    } finally {
      converting.value = false
    }
  }

  const onImagePick = (f) => {
    imageFile.value = f?.raw || null
  }

  const convertImage = async () => {
    if (!imageFile.value) return
    if (imageTarget.value === 'webp-local') {
      const url = URL.createObjectURL(imageFile.value)
      const img = new Image()
      img.onload = () => {
        const canvas = document.createElement('canvas')
        canvas.width = img.naturalWidth
        canvas.height = img.naturalHeight
        canvas.getContext('2d').drawImage(img, 0, 0)
        canvas.toBlob((blob) => {
          if (!blob) return ElMessage.error('转换失败')
          downloadBlob(blob, 'converted.webp')
          ElMessage.success('WebP 已下载（本地转换）')
        }, 'image/webp', 0.92)
      }
      img.src = url
      return
    }
    const fd = new FormData()
    fd.append('file', imageFile.value)
    fd.append('target', imageTarget.value)
    const ext = imageTarget.value === 'jpeg' ? 'jpg' : 'png'
    await postConvert(fd, `converted.${ext}`)
  }

  const onOfficePick = (f) => {
    officeFile.value = f?.raw || null
  }

  const convertOffice = async () => {
    if (!officeFile.value) return
    const fd = new FormData()
    fd.append('file', officeFile.value)
    fd.append('target', 'pdf')
    const base = officeFile.value.name.replace(/\.[^.]+$/, '')
    await postConvert(fd, `${base}.pdf`)
  }

  const onImgPdfPick = (f) => {
    imgPdfFile.value = f?.raw || null
  }

  const convertImgPdf = async () => {
    if (!imgPdfFile.value) return
    const fd = new FormData()
    fd.append('file', imgPdfFile.value)
    fd.append('target', 'pdf')
    const base = imgPdfFile.value.name.replace(/\.[^.]+$/, '')
    await postConvert(fd, `${base}.pdf`)
  }

  const convertTextPdf = async () => {
    const fd = new FormData()
    fd.append('text', textContent.value)
    fd.append('target', 'pdf')
    await postConvert(fd, 'export.pdf')
  }

  onMounted(async () => {
    try {
      const res = await getOfficeConvertCapabilities()
      caps.value = res.data || null
    } catch {
      caps.value = { libreOffice: false, maxUploadMB: 20 }
    }
  })
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .hint { font-size: 13px; color: #909399; margin: 0 0 12px; }
  .mt-3 { margin-top: 12px; }
  .mb-3 { margin-bottom: 12px; }
  .file-name { margin-top: 8px; font-size: 13px; color: #606266; }
</style>
