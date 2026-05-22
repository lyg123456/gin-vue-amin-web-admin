<template>
  <div class="tool-panel">
    <el-alert v-if="caps" :title="caps.hint" :type="caps.ffmpeg ? 'success' : 'warning'" show-icon :closable="false" class="mb-3" />
    <el-tabs v-model="tab">
      <el-tab-pane label="提取音频" name="extract_audio">
        <p class="hint">从视频中提取 MP3 音频（需服务器 FFmpeg）</p>
        <el-upload :auto-upload="false" :limit="1" accept="video/*" @change="(f) => setVideo(f)">
          <el-button type="primary">选择视频</el-button>
        </el-upload>
        <el-button class="mt-3" type="primary" :loading="loading" :disabled="!videoFile || !caps?.ffmpeg" @click="runVideo('extract_audio')">提取并下载</el-button>
      </el-tab-pane>
      <el-tab-pane label="去除音频" name="remove_audio">
        <p class="hint">生成无音轨视频（需 FFmpeg）</p>
        <el-upload :auto-upload="false" :limit="1" accept="video/*" @change="(f) => setVideo(f)">
          <el-button>选择视频</el-button>
        </el-upload>
        <el-button class="mt-3" type="primary" :loading="loading" :disabled="!videoFile || !caps?.ffmpeg" @click="runVideo('remove_audio')">处理并下载</el-button>
      </el-tab-pane>
      <el-tab-pane label="提取画面/背景帧" name="extract_frame">
        <p class="hint">从视频截取首帧为 JPG，或从图片采样生成纯色背景图</p>
        <el-radio-group v-model="frameMode" class="mb-2">
          <el-radio-button label="video">视频截帧</el-radio-button>
          <el-radio-button label="image">图片采样背景</el-radio-button>
        </el-radio-group>
        <el-upload :auto-upload="false" :limit="1" :accept="frameMode === 'video' ? 'video/*' : 'image/*'" @change="(f) => setFrameFile(f)">
          <el-button>选择文件</el-button>
        </el-upload>
        <el-button class="mt-3" type="primary" :loading="loading" :disabled="!frameFile" @click="runFrame">导出</el-button>
      </el-tab-pane>
      <el-tab-pane label="图片合成" name="composite">
        <p class="hint">上传背景图 + 前景图，叠加后下载</p>
        <el-row :gutter="12">
          <el-col :span="12">
            <span class="label">背景图</span>
            <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" @change="(f) => (bgFile = f?.raw)">
              <el-button size="small">选择背景</el-button>
            </el-upload>
          </el-col>
          <el-col :span="12">
            <span class="label">前景图</span>
            <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" @change="(f) => (fgFile = f?.raw)">
              <el-button size="small">选择前景</el-button>
            </el-upload>
          </el-col>
        </el-row>
        <el-row :gutter="12" class="mt-2">
          <el-col :span="8"><span class="label">X</span><el-input-number v-model="posX" :min="0" class="w-full" /></el-col>
          <el-col :span="8"><span class="label">Y</span><el-input-number v-model="posY" :min="0" class="w-full" /></el-col>
          <el-col :span="8"><span class="label">缩放%</span><el-input-number v-model="scale" :min="10" :max="200" class="w-full" /></el-col>
        </el-row>
        <el-button class="mt-3" type="primary" :loading="loading" :disabled="!bgFile || !fgFile" @click="runComposite">合成并下载</el-button>
      </el-tab-pane>
      <el-tab-pane label="本地抠图" name="chroma">
        <p class="hint">浏览器本地绿幕抠图（不上传服务器），适合纯色背景视频截图</p>
        <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" @change="onChromaPick">
          <el-button>选择图片</el-button>
        </el-upload>
        <el-color-picker v-model="chromaColor" class="mt-2" />
        <span class="hint ml-2">抠除颜色</span>
        <el-slider v-model="chromaTolerance" :min="5" :max="80" class="mt-2" />
        <el-button class="mt-3" type="primary" :disabled="!chromaUrl" @click="downloadChroma">下载透明 PNG</el-button>
        <img v-if="chromaUrl" :src="chromaPreview" class="preview mt-2" alt="preview" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    compositeImages,
    extractImageBackground,
    getMediaCapabilities,
    processMediaVideo
  } from '@/api/publicPortalOffice'

  const tab = ref('extract_audio')
  const caps = ref(null)
  const loading = ref(false)
  const videoFile = ref(null)
  const frameFile = ref(null)
  const frameMode = ref('video')
  const bgFile = ref(null)
  const fgFile = ref(null)
  const posX = ref(0)
  const posY = ref(0)
  const scale = ref(100)
  const chromaColor = ref('#00ff00')
  const chromaTolerance = ref(30)
  const chromaUrl = ref('')
  const chromaPreview = ref('')

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
    a.download = name || 'download'
    a.click()
    URL.revokeObjectURL(a.href)
    ElMessage.success('已开始下载')
  }

  const setVideo = (f) => {
    videoFile.value = f?.raw || null
  }
  const setFrameFile = (f) => {
    frameFile.value = f?.raw || null
  }

  const runVideo = async (action) => {
    if (!videoFile.value) return
    loading.value = true
    try {
      const fd = new FormData()
      fd.append('file', videoFile.value)
      fd.append('action', action)
      const res = await processMediaVideo(fd)
      downloadBlob(res, action === 'extract_audio' ? 'audio.mp3' : 'silent.mp4')
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '处理失败')
    } finally {
      loading.value = false
    }
  }

  const runFrame = async () => {
    if (!frameFile.value) return
    loading.value = true
    try {
      if (frameMode.value === 'video') {
        const fd = new FormData()
        fd.append('file', frameFile.value)
        fd.append('action', 'extract_frame')
        const res = await processMediaVideo(fd)
        downloadBlob(res, 'frame.jpg')
      } else {
        const fd = new FormData()
        fd.append('file', frameFile.value)
        const res = await extractImageBackground(fd)
        downloadBlob(res, 'background.png')
      }
    } catch (e) {
      ElMessage.error(e?.message || '失败')
    } finally {
      loading.value = false
    }
  }

  const runComposite = async () => {
    loading.value = true
    try {
      const fd = new FormData()
      fd.append('background', bgFile.value)
      fd.append('foreground', fgFile.value)
      fd.append('x', String(posX.value))
      fd.append('y', String(posY.value))
      fd.append('scale', String(scale.value))
      const res = await compositeImages(fd)
      downloadBlob(res, 'merged.jpg')
    } catch (e) {
      ElMessage.error(e?.message || '合成失败')
    } finally {
      loading.value = false
    }
  }

  const onChromaPick = (f) => {
    const file = f?.raw
    if (!file) return
    if (chromaUrl.value) URL.revokeObjectURL(chromaUrl.value)
    chromaUrl.value = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = img.naturalWidth
      canvas.height = img.naturalHeight
      const ctx = canvas.getContext('2d')
      ctx.drawImage(img, 0, 0)
      const hex = chromaColor.value.replace('#', '')
      const tr = parseInt(hex.slice(0, 2), 16)
      const tg = parseInt(hex.slice(2, 4), 16)
      const tb = parseInt(hex.slice(4, 6), 16)
      const tol = chromaTolerance.value
      const data = ctx.getImageData(0, 0, canvas.width, canvas.height)
      for (let i = 0; i < data.data.length; i += 4) {
        const dr = Math.abs(data.data[i] - tr)
        const dg = Math.abs(data.data[i + 1] - tg)
        const db = Math.abs(data.data[i + 2] - tb)
        if (dr < tol && dg < tol && db < tol) data.data[i + 3] = 0
      }
      ctx.putImageData(data, 0, 0)
      chromaPreview.value = canvas.toDataURL('image/png')
    }
    img.src = chromaUrl.value
  }

  const downloadChroma = () => {
    if (!chromaPreview.value) return
    const a = document.createElement('a')
    a.href = chromaPreview.value
    a.download = 'cutout.png'
    a.click()
  }

  onMounted(async () => {
    try {
      const res = await getMediaCapabilities()
      caps.value = res.data || null
    } catch {
      caps.value = { ffmpeg: false, hint: '无法获取能力信息' }
    }
  })
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .hint { font-size: 13px; color: #909399; margin: 0 0 8px; }
  .mt-2 { margin-top: 8px; }
  .mt-3 { margin-top: 12px; }
  .mb-2 { margin-bottom: 8px; }
  .mb-3 { margin-bottom: 12px; }
  .ml-2 { margin-left: 8px; }
  .label { font-size: 12px; color: #606266; display: block; margin-bottom: 4px; }
  .preview { max-width: 100%; max-height: 200px; border-radius: 8px; }
  .w-full { width: 100%; }
</style>
