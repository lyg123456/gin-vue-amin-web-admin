<template>
  <div class="gva-table-box">
    <el-alert
      title="脚本：volc-ark.api-key；成片推荐 dashscope-video.api-key（阿里云 video-synthesis，见 config.yaml）。有素材图时用 i2v 模型；时长 30～120 秒会映射为 5/10 秒档位。"
      type="info"
      show-icon
      :closable="false"
      class="mb-4"
    />
    <el-form :model="form" label-width="120px" class="max-w-3xl">
      <el-form-item label="标题" required>
        <el-input v-model="form.title" maxlength="200" show-word-limit />
      </el-form-item>
      <el-form-item label="创意文字">
        <el-input v-model="form.promptText" type="textarea" :rows="3" placeholder="口播要点、卖点、受众等" />
      </el-form-item>
      <el-form-item label="场景描述">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="产品/场景/风格说明" />
      </el-form-item>
      <el-form-item label="目标时长(秒)">
        <el-slider v-model="form.durationSec" :min="30" :max="120" :step="5" show-input />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="scriptLoading" @click="onGenScript">生成脚本</el-button>
        <el-button type="success" :loading="saveLoading" :disabled="!form.script?.trim()" @click="onSave">
          入库并可选生成成片
        </el-button>
        <el-button @click="$router.push({ name: 'shortVideoList' })">短视频列表</el-button>
      </el-form-item>
      <el-form-item label="脚本">
        <el-input v-model="form.script" type="textarea" :rows="12" placeholder="生成后在此编辑" />
      </el-form-item>
      <el-form-item label="封面图">
        <SelectImage v-model="coverImage" file-type="image" />
      </el-form-item>
      <el-form-item label="首帧图" required>
        <SelectImage v-model="firstFrameUrl" file-type="image" />
        <p class="tip">支持 JPG/PNG/BMP/WEBP；本地上传会自动转 Base64 提交，或使用公网 HTTPS 图片链接</p>
      </el-form-item>
      <el-form-item label="尾帧图" required>
        <SelectImage v-model="lastFrameUrl" file-type="image" />
        <p class="tip">须与首帧同为可识别图片格式，勿用 GIF/SVG</p>
      </el-form-item>
      <el-form-item label="成片视频">
        <SelectImage v-model="videoUrl" file-type="video" />
        <p class="tip">未配置视频 API 时可先手动上传</p>
      </el-form-item>
      <el-form-item label="提交成片任务">
        <el-switch v-model="form.autoGenerate" />
        <span class="tip ml-2">入库后调用 DashScope 异步成片（需 dashscope-video.api-key）</span>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import { generateShortVideoScript, createShortVideoWithAI } from '@/api/contentShortVideo'
  import { parseAiArticleOutput } from '@/utils/parseAiArticleOutput'

  defineOptions({ name: 'ShortVideoAiGenerate' })

  const router = useRouter()
  const scriptLoading = ref(false)
  const saveLoading = ref(false)
  const firstFrameUrl = ref('')
  const lastFrameUrl = ref('')
  const coverImage = ref('')
  const videoUrl = ref('')

  const pickImageUrl = (val) => {
    if (Array.isArray(val)) return (val[0] || '').trim()
    return (val || '').trim()
  }

  const form = reactive({
    title: '',
    promptText: '',
    description: '',
    durationSec: 60,
    script: '',
    autoGenerate: false
  })

  const onGenScript = async () => {
    if (!form.title?.trim()) {
      ElMessage.warning('请填写标题')
      return
    }
    scriptLoading.value = true
    try {
      const res = await generateShortVideoScript({
        title: form.title.trim(),
        promptText: form.promptText.trim(),
        description: form.description.trim(),
        durationSec: form.durationSec
      })
      if (res.code === 0 && res.data?.script) {
        const parsed = parseAiArticleOutput(res.data.script)
        form.script = parsed.content || res.data.script
        ElMessage.success('脚本已生成')
      }
    } finally {
      scriptLoading.value = false
    }
  }

  const onSave = async () => {
    if (!form.title?.trim() || !form.script?.trim()) {
      ElMessage.warning('请填写标题并生成脚本')
      return
    }
    const first = pickImageUrl(firstFrameUrl.value)
    const last = pickImageUrl(lastFrameUrl.value)
    if (form.autoGenerate && (!first || !last)) {
      ElMessage.warning('自动生成成片需同时上传首帧图与尾帧图')
      return
    }
    saveLoading.value = true
    try {
      const cover = pickImageUrl(coverImage.value)
      const video = pickImageUrl(videoUrl.value)
      const res = await createShortVideoWithAI({
        title: form.title.trim(),
        promptText: form.promptText.trim(),
        description: form.description.trim(),
        durationSec: form.durationSec,
        script: form.script,
        coverImage: cover,
        firstFrameUrl: first,
        lastFrameUrl: last,
        videoUrl: video,
        autoGenerate: form.autoGenerate
      })
      if (res.code === 0 && res.data) {
        ElMessage.success(res.msg || '已入库')
        router.push({ name: 'shortVideoList' })
      }
    } finally {
      saveLoading.value = false
    }
  }
</script>

<style scoped>
  .mb-4 { margin-bottom: 16px; }
  .max-w-3xl { max-width: 900px; }
  .tip { font-size: 12px; color: #909399; margin: 6px 0 0; }
  .ml-2 { margin-left: 8px; }
</style>
