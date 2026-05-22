<template>
  <div class="tool-panel">
    <el-row :gutter="16">
      <el-col :xs="24" :md="10">
        <h4 class="sub-title">排版模板</h4>
        <div class="tpl-grid">
          <el-button
            v-for="t in templates"
            :key="t.id"
            :type="activeTpl === t.id ? 'primary' : 'default'"
            class="tpl-btn"
            @click="applyTemplate(t)"
          >
            {{ t.name }}
          </el-button>
        </div>
        <h4 class="sub-title mt-3">AI 改写风格</h4>
        <el-select v-model="aiStyle" class="w-full" placeholder="选择风格">
          <el-option v-for="(label, key) in styles" :key="key" :label="label" :value="key" />
        </el-select>
        <el-button type="primary" class="mt-2 w-full" :loading="aiLoading" @click="runAiRewrite">AI 改写</el-button>
        <p class="hint">需配置 volc-ark 或 baidu-wenxin（与后台 AI 写文章相同）</p>
      </el-col>
      <el-col :xs="24" :md="14">
        <h4 class="sub-title">推文内容</h4>
        <el-input v-model="rawText" type="textarea" :rows="8" placeholder="输入推文原文…" />
        <el-button class="mt-2" @click="beautify">一键美化排版</el-button>
        <el-button class="mt-2 ml-2" @click="copyOut">复制结果</el-button>
        <div class="preview-box mt-3" :style="previewStyle">
          <div class="preview-inner" v-html="previewHtml" />
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { getTweetStyles, rewriteTweet } from '@/api/publicPortalOffice'

  const rawText = ref('分享一条今日见闻：\n\n坚持复盘，比盲目加班更重要。\n\n#成长 #职场')
  const activeTpl = ref('classic')
  const aiStyle = ref('default')
  const styles = ref({})
  const aiLoading = ref(false)

  const templates = [
    { id: 'classic', name: '经典分段', style: { fontFamily: 'Georgia, serif', lineHeight: '1.8', padding: '20px', background: '#fafafa', borderRadius: '12px' } },
    { id: 'xhs', name: '小红书', style: { fontFamily: 'sans-serif', lineHeight: '1.9', padding: '24px', background: 'linear-gradient(135deg,#fff5f5,#fff)', borderRadius: '16px', color: '#333' } },
    { id: 'quote', name: '金句突出', style: { fontFamily: 'serif', lineHeight: '2', padding: '28px', background: '#1a1a2e', color: '#eee', borderRadius: '8px' } },
    { id: 'list', name: '清单体', style: { fontFamily: 'sans-serif', lineHeight: '1.7', padding: '20px', background: '#f0f9ff', borderLeft: '4px solid #409eff' } },
    { id: 'minimal', name: '极简白', style: { fontFamily: 'system-ui', lineHeight: '1.6', padding: '16px', background: '#fff', border: '1px solid #e5e7eb' } },
    { id: 'warm', name: '暖色推文', style: { fontFamily: 'sans-serif', lineHeight: '1.85', padding: '22px', background: '#fffbeb', color: '#78350f', borderRadius: '12px' } }
  ]

  const previewStyle = computed(() => {
    const t = templates.find((x) => x.id === activeTpl.value)
    return t?.style || {}
  })

  const previewHtml = computed(() => {
    const t = rawText.value.replace(/</g, '&lt;').replace(/\n/g, '<br/>')
    if (activeTpl.value === 'quote') {
      const lines = rawText.value.split('\n').filter(Boolean)
      const first = lines[0] || ''
      const rest = lines.slice(1).join('<br/>')
      return `<p style="font-size:1.35em;margin:0 0 12px;">${first.replace(/</g, '&lt;')}</p><p style="opacity:.85">${rest.replace(/</g, '&lt;')}</p>`
    }
    if (activeTpl.value === 'list') {
      return rawText.value
        .split('\n')
        .filter(Boolean)
        .map((l) => `• ${l.replace(/</g, '&lt;')}`)
        .join('<br/>')
    }
    return t
  })

  const applyTemplate = (t) => {
    activeTpl.value = t.id
    ElMessage.success(`已应用模板：${t.name}`)
  }

  const beautify = () => {
    let t = rawText.value.trim()
    t = t.replace(/\r\n/g, '\n')
    t = t.replace(/([。！？])\s*/g, '$1\n\n')
    t = t.replace(/\n{3,}/g, '\n\n')
    if (!t.includes('#') && t.length > 20) t += '\n\n#推文'
    rawText.value = t.trim()
    ElMessage.success('已美化分段与标点')
  }

  const runAiRewrite = async () => {
    aiLoading.value = true
    try {
      const res = await rewriteTweet({ text: rawText.value, style: aiStyle.value })
      if (res.data?.text) {
        rawText.value = res.data.text
        ElMessage.success('AI 改写完成')
      }
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || 'AI 改写失败')
    } finally {
      aiLoading.value = false
    }
  }

  const copyOut = async () => {
    await navigator.clipboard.writeText(rawText.value)
    ElMessage.success('已复制纯文本')
  }

  onMounted(async () => {
    try {
      const res = await getTweetStyles()
      styles.value = res.data || {}
    } catch {
      styles.value = { default: '默认' }
    }
  })
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .sub-title { margin: 0 0 10px; font-size: 14px; font-weight: 600; color: #303133; }
  .tpl-grid { display: flex; flex-wrap: wrap; gap: 8px; }
  .tpl-btn { margin: 0; }
  .hint { font-size: 12px; color: #909399; margin-top: 8px; }
  .mt-2 { margin-top: 8px; }
  .mt-3 { margin-top: 12px; }
  .ml-2 { margin-left: 8px; }
  .w-full { width: 100%; }
  .preview-box { min-height: 160px; }
  .preview-inner { word-break: break-word; }
</style>
