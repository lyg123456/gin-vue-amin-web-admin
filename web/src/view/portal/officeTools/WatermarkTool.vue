<template>
  <div class="tool-panel wm-tool">
    <el-alert v-if="caps" :title="caps.hint" :type="caps.ffmpeg ? 'info' : 'warning'" show-icon :closable="false" class="mb-3" />

    <el-form label-position="top" class="wm-form">
      <el-row :gutter="16">
        <el-col :span="12" :xs="24">
          <el-form-item label="平台预设">
            <el-select v-model="preset" class="w-full" @change="onPresetChange">
              <el-option
                v-for="p in caps?.presets || []"
                :key="p.id"
                :label="`${p.name} — ${p.platform}`"
                :value="p.id"
              />
            </el-select>
            <p v-if="currentPreset" class="preset-desc">{{ currentPreset.description }}</p>
          </el-form-item>
        </el-col>
        <el-col :span="12" :xs="24">
          <el-form-item label="处理方式">
            <el-radio-group v-model="method">
              <el-radio
                v-for="m in caps?.methods || []"
                :key="m.id"
                :label="m.id"
                :disabled="m.id === 'delogo' && !isVideo && !caps?.ffmpeg"
              >
                {{ m.name }}
              </el-radio>
            </el-radio-group>
            <p class="preset-desc">{{ methodDesc }}</p>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item v-if="preset === 'custom'" label="自定义区域（占画面百分比 %）">
        <el-row :gutter="8">
          <el-col :span="6"><span class="dim">X</span><el-input-number v-model="custom.x" :min="0" :max="95" class="w-full" /></el-col>
          <el-col :span="6"><span class="dim">Y</span><el-input-number v-model="custom.y" :min="0" :max="95" class="w-full" /></el-col>
          <el-col :span="6"><span class="dim">宽</span><el-input-number v-model="custom.w" :min="5" :max="100" class="w-full" /></el-col>
          <el-col :span="6"><span class="dim">高</span><el-input-number v-model="custom.h" :min="5" :max="100" class="w-full" /></el-col>
        </el-row>
      </el-form-item>

      <el-form-item label="上传图片或视频">
        <el-upload
          :auto-upload="false"
          :limit="1"
          accept="image/*,video/*"
          :on-change="onFileChange"
          :on-remove="() => (file = null)"
        >
          <el-button type="primary">选择文件</el-button>
        </el-upload>
        <p class="preset-desc">
          支持 JPG/PNG/WebP、MP4/MOV。请确保<b>橙色框完全盖住水印文字</b>（含「即梦AI」等），可选用「即梦AI」预设 +「智能填充」。
        </p>
      </el-form-item>

      <div v-if="previewUrl && !isVideo" class="preview-wrap">
        <img :src="previewUrl" class="preview-img" alt="preview" />
        <div
          v-if="overlayStyle"
          class="wm-overlay"
          :style="overlayStyle"
          title="预估去水印区域"
        />
      </div>

      <el-button type="primary" :loading="loading" :disabled="!file" @click="run">处理并下载</el-button>
    </el-form>

    <el-collapse class="mt-3">
      <el-collapse-item title="各平台水印位置说明" name="help">
        <ul class="help-list">
          <li><b>即梦AI / AI底栏</b>：选对应预设 +「智能填充」，框住整行文字与图标。</li>
          <li><b>抖音 / TikTok / 快手</b>：右下账号，建议「智能填充」或 Delogo（视频）。</li>
          <li><b>小红书</b>：底部整条水印，建议「裁剪去除」。</li>
          <li><b>B站</b>：右上 UP 主标识，建议「区域模糊」。</li>
          <li><b>视频号 / Instagram</b>：顶部或右下标，按预览框微调自定义区域。</li>
          <li>动态游走水印、全屏半透明大字无法完美去除，仅供固定角标场景。</li>
        </ul>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getWatermarkCapabilities, removeWatermark } from '@/api/publicPortalOffice'

  const caps = ref(null)
  const preset = ref('jimeng')
  const method = ref('fill')
  const file = ref(null)
  const loading = ref(false)
  const previewUrl = ref('')
  const custom = reactive({ x: 75, y: 85, w: 20, h: 10 })

  const isVideo = computed(() => {
    const n = file.value?.name || ''
    return /\.(mp4|mov|webm|mkv|avi)$/i.test(n)
  })

  const currentPreset = computed(() => caps.value?.presets?.find((p) => p.id === preset.value))

  const methodDesc = computed(() => {
    const m = caps.value?.methods?.find((x) => x.id === method.value)
    return m?.desc || ''
  })

  const overlayStyle = computed(() => {
    const p = preset.value === 'custom' ? custom : currentPreset.value
    if (!p) return null
    const pad = 18
    const x = Math.max(0, p.x - pad * 0.5)
    const y = Math.max(0, p.y - pad * 0.4)
    const w = Math.min(100 - x, p.w + pad)
    const h = Math.min(100 - y, p.h + pad)
    return {
      left: `${x}%`,
      top: `${y}%`,
      width: `${w}%`,
      height: `${h}%`
    }
  })

  const onPresetChange = () => {
    const p = currentPreset.value
    if (p && preset.value === 'custom') {
      custom.x = p.x
      custom.y = p.y
      custom.w = p.w
      custom.h = p.h
    }
    if (preset.value === 'xiaohongshu' && !isVideo.value) method.value = 'crop'
  }

  const onFileChange = (f) => {
    file.value = f?.raw || null
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
    const video = /\.(mp4|mov|webm|mkv|avi)$/i.test(file.value?.name || '')
    if (file.value && !video) {
      previewUrl.value = URL.createObjectURL(file.value)
    }
    if (video) {
      if (method.value === 'crop') method.value = 'delogo'
      if (!caps.value?.ffmpeg) ElMessage.warning('视频去水印需要服务器安装 FFmpeg')
    }
  }

  const downloadBlob = (res, name) => {
    const blob = res?.data instanceof Blob ? res.data : res
    if (blob?.type?.includes('json')) {
      blob.text().then((t) => {
        try {
          ElMessage.error(JSON.parse(t).msg || '处理失败')
        } catch {
          ElMessage.error('处理失败')
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

  const run = async () => {
    if (!file.value) return
    const fd = new FormData()
    fd.append('file', file.value)
    fd.append('preset', preset.value)
    fd.append('method', isVideo.value && method.value !== 'delogo' ? 'delogo' : method.value)
    if (preset.value === 'custom') {
      fd.append('x', String(custom.x))
      fd.append('y', String(custom.y))
      fd.append('w', String(custom.w))
      fd.append('h', String(custom.h))
    }
    loading.value = true
    try {
      const res = await removeWatermark(fd)
      downloadBlob(res, file.value.name.replace(/(\.[^.]+)$/, '_nowm$1'))
    } catch (e) {
      ElMessage.error(e?.message || '失败')
    } finally {
      loading.value = false
    }
  }

  onMounted(async () => {
    const res = await getWatermarkCapabilities()
    caps.value = res.data || null
  })
</script>

<style scoped>
  .wm-form {
    max-width: 900px;
  }
  .preset-desc {
    margin: 6px 0 0;
    font-size: 13px;
    color: #888;
    line-height: 1.5;
  }
  .preview-wrap {
    position: relative;
    display: inline-block;
    max-width: 100%;
    margin-bottom: 14px;
  }
  .preview-img {
    max-width: 100%;
    max-height: 320px;
    display: block;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
  }
  .wm-overlay {
    position: absolute;
    box-sizing: border-box;
    border: 2px dashed #ff6600;
    background: rgba(255, 102, 0, 0.15);
    pointer-events: none;
  }
  .help-list {
    margin: 0;
    padding-left: 1.2em;
    font-size: 14px;
    color: #555;
    line-height: 1.7;
  }
  .dim {
    font-size: 12px;
    color: #888;
  }
  .mb-3 {
    margin-bottom: 12px;
  }
  .mt-3 {
    margin-top: 14px;
  }
  .w-full {
    width: 100%;
  }
</style>
