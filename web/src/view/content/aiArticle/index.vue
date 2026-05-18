<template>
  <div class="ai-article-page">
    <el-alert
      title="推理接入点」的 Endpoint ID（ep- 开头）"
      type="info"
      show-icon
      :closable="false"
      class="mb-4"
    />
    <div class="gva-table-box">
      <el-form :model="form" label-width="108px" class="max-w-3xl">
        <el-form-item label="规则文件">
          <el-upload
            ref="rulesUploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="onRulesFileChange"
            :on-remove="onRulesFileRemove"
            accept=".txt,.md,.markdown,.text,.json,.csv,.log"
          >
            <template #trigger>
              <el-button type="primary">选择规则文件</el-button>
            </template>
            <template #tip>
              <span class="upload-tip">本地解析为 UTF-8 文本，建议 ≤512KB；与下方「写作要求」可同时使用（文件优先）。</span>
            </template>
          </el-upload>
          <p v-if="rulesFileName" class="file-meta">
            已加载：<strong>{{ rulesFileName }}</strong> · 约 <strong>{{ rulesFileCharCount }}</strong> 字（Unicode）
          </p>
        </el-form-item>
        <el-form-item label="核心标题" required>
          <el-input v-model="form.title" maxlength="200" show-word-limit placeholder="必填；将写入文章标题并作为生成提示的核心标题" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="form.keywords"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
            placeholder="可选，多个用逗号分隔；生成后会写入文章「SEO 关键词」"
          />
        </el-form-item>
        <el-form-item label="写作要求">
          <el-input
            v-model="form.rules"
            type="textarea"
            :rows="3"
            maxlength="2000"
            show-word-limit
            placeholder="可选；上传规则文件后，此处可作为补充说明（仍以文件规则为最高优先级）"
          />
        </el-form-item>
        <el-form-item label="目标字数">
          <el-input-number v-model="form.wordCount" :min="400" :max="8000" :step="100" controls-position="right" />
        </el-form-item>
        <el-form-item label="随机度">
          <el-slider v-model="form.temperature" :min="0.1" :max="1" :step="0.1" show-input />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="onGenerate">生成正文</el-button>
          <el-button :loading="diagnoseLoading" @click="onDiagnose">诊断连接</el-button>
          <el-button :disabled="!syncForm.content?.trim()" @click="copyContent">复制正文</el-button>
          <el-button type="success" :disabled="!syncForm.content?.trim() || !syncForm.title?.trim()" @click="goCreateArticle">
            填入「文章管理」新建
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider content-position="left">与文章管理新建表单一致（可编辑后跳转）</el-divider>
      <el-form :model="syncForm" label-width="110px" class="max-w-3xl sync-form">
        <el-form-item label="标题">
          <el-input v-model="syncForm.title" maxlength="200" show-word-limit placeholder="与文章管理「标题」一致" />
        </el-form-item>
        <el-form-item label="Slug">
          <div class="flex gap-2 w-full">
            <el-input v-model="syncForm.slug" placeholder="URL 路径，如 how-to-seo；留空则在文章管理内自动生成" />
            <el-button @click="regenSlug">按标题生成</el-button>
          </div>
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="syncForm.summary" type="textarea" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="分类">
          <el-cascader
            v-model="categoryPick"
            :options="categoryTreeOpts"
            :props="cascaderProps"
            clearable
            filterable
            placeholder="与文章管理一致，可选一级或二级"
            style="width: 100%"
            @change="onCategoryCascaderChange"
          />
        </el-form-item>
        <el-form-item label="配图">
          <SelectImage v-model="galleryUrls" multiple :max-update-count="6" file-type="image" />
          <p class="field-tip">最多 6 张，对应文章 <code>cover_image</code>（逗号分隔）</p>
        </el-form-item>
        <el-form-item label="SEO 标题">
          <el-input v-model="syncForm.seoTitle" maxlength="200" show-word-limit placeholder="默认同标题，可改" />
        </el-form-item>
        <el-form-item label="SEO 关键词">
          <el-input v-model="syncForm.seoKeywords" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="SEO 描述">
          <el-input v-model="syncForm.seoDescription" type="textarea" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="正文(Markdown)">
          <el-input v-model="syncForm.content" type="textarea" :rows="14" placeholder="生成后出现在此，也可直接粘贴修改" />
        </el-form-item>
      </el-form>
    </div>

    <el-dialog v-model="diagnoseVisible" title="AI 连接诊断" width="640px" destroy-on-close>
      <pre class="diagnose-pre">{{ diagnoseText }}</pre>
      <template #footer>
        <el-button type="primary" @click="copyDiagnose">复制 JSON</el-button>
        <el-button @click="diagnoseVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { generateArticleByBaidu, diagnoseBaiduWenxin } from '@/api/contentArticle'
  import { getContentArticleCategoryTree } from '@/api/contentArticleCategory'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import { parseAiArticleOutput, buildSummaryFromMarkdown } from '@/utils/parseAiArticleOutput'

  defineOptions({
    name: 'ContentAiArticle'
  })

  const AI_PREFILL_KEY = 'GVA_CONTENT_ARTICLE_AI_PREFILL'

  const router = useRouter()
  const loading = ref(false)
  const diagnoseLoading = ref(false)
  const diagnoseVisible = ref(false)
  const diagnoseText = ref('')
  const rulesUploadRef = ref()
  const rulesFileText = ref('')
  const rulesFileName = ref('')

  const rulesFileCharCount = computed(() => {
    const s = rulesFileText.value || ''
    return [...s].length
  })

  const form = reactive({
    title: '',
    keywords: '',
    rules: '',
    wordCount: 1000,
    temperature: 0.7
  })

  /** 与文章管理抽屉字段一一对应 */
  const syncForm = reactive({
    title: '',
    slug: '',
    summary: '',
    content: '',
    seoTitle: '',
    seoKeywords: '',
    seoDescription: '',
    categoryId: 0
  })

  const galleryUrls = ref([])

  const cascaderProps = {
    value: 'value',
    label: 'label',
    children: 'children',
    checkStrictly: true,
    emitPath: false
  }
  const categoryTreeOpts = ref([])
  const categoryPick = ref()

  const loadCategoryTree = async () => {
    const res = await getContentArticleCategoryTree()
    if (res.code === 0) {
      categoryTreeOpts.value = res.data?.list || []
    }
  }

  const onCategoryCascaderChange = (v) => {
    syncForm.categoryId = v ? Number(v) : 0
  }

  const syncCategoryPickFromSync = () => {
    const id = Number(syncForm.categoryId || 0)
    categoryPick.value = id > 0 ? id : undefined
  }

  onMounted(() => {
    loadCategoryTree()
  })

  const onRulesFileChange = (uploadFile) => {
    const raw = uploadFile.raw
    if (!raw) return
    const maxBytes = 512 * 1024
    if (raw.size > maxBytes) {
      ElMessage.warning('文件过大，请控制在 512KB 以内')
      rulesUploadRef.value?.clearFiles()
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      rulesFileText.value = String(reader.result || '')
      rulesFileName.value = raw.name
      ElMessage.success('已读取规则文件')
    }
    reader.onerror = () => {
      ElMessage.error('读取文件失败')
      rulesUploadRef.value?.clearFiles()
    }
    reader.readAsText(raw, 'UTF-8')
  }

  const onRulesFileRemove = () => {
    rulesFileText.value = ''
    rulesFileName.value = ''
  }

  /** 将接口原始全文解析后写入各字段：正文仅保留正文区，摘要/SEO 仅保留「标签：」后内容 */
  const applyParsedToSyncForm = (rawContent) => {
    const parsed = parseAiArticleOutput(rawContent)
    const title = form.title.trim().slice(0, 200)

    syncForm.content = parsed.content
    syncForm.title = title
    if (!String(syncForm.slug || '').trim()) {
      syncForm.slug = suggestSlug(title)
    }

    syncForm.summary = (parsed.summary || buildSummaryFromMarkdown(parsed.content)).slice(0, 500)
    syncForm.seoTitle = (parsed.seoTitle || title).slice(0, 200)
    syncForm.seoDescription = (parsed.seoDescription || '').slice(0, 500)
    const kw = parsed.seoKeywords || form.keywords.trim()
    syncForm.seoKeywords = kw.slice(0, 500)
  }

  /** 英文 Slug；纯中文等无法生成时用 article-时间戳 */
  const suggestSlug = (title) => {
    const raw = String(title || '').trim().toLowerCase()
    const x = raw
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 120)
    if (x.length >= 3) return x
    return `article-${Date.now()}`
  }

  const regenSlug = () => {
    const t = syncForm.title?.trim() || form.title?.trim()
    if (!t) {
      ElMessage.warning('请先填写标题')
      return
    }
    syncForm.slug = suggestSlug(t)
    ElMessage.success('已生成 Slug')
  }

  const onGenerate = async () => {
    if (!form.title?.trim()) {
      ElMessage.warning('请先填写文章标题')
      return
    }
    loading.value = true
    try {
      const res = await generateArticleByBaidu({
        title: form.title.trim(),
        keywords: form.keywords.trim(),
        rules: form.rules.trim(),
        rulesFileContent: rulesFileText.value.trim(),
        wordCount: form.wordCount,
        temperature: form.temperature
      })
      if (res.code === 0 && res.data?.content) {
        applyParsedToSyncForm(String(res.data.content || ''))
        syncCategoryPickFromSync()
        ElMessage.success(res.msg || '生成成功，已同步到下方文章字段，可修改后跳转文章管理保存')
      }
    } finally {
      loading.value = false
    }
  }

  const onDiagnose = async () => {
    diagnoseLoading.value = true
    try {
      const res = await diagnoseBaiduWenxin()
      if (res.code === 0 && res.data) {
        diagnoseText.value = JSON.stringify(res.data, null, 2)
        diagnoseVisible.value = true
        const d = res.data
        if (d.chatOk) {
          const p = d.provider || ''
          if (p === 'volc-ark') {
            ElMessage.success('诊断：火山方舟（豆包）调用正常')
          } else if (p === 'baidu-v2' || d.authMode === 'v2-bearer') {
            ElMessage.success('诊断：百度千帆 V2 Bearer 调用正常')
          } else if (p === 'baidu-oauth' || d.authMode === 'oauth-access-token') {
            ElMessage.success('诊断：百度 OAuth 与模型调用正常')
          } else {
            ElMessage.success('诊断：模型调用正常')
          }
        } else if (!d.tokenOk) {
          ElMessage.warning('诊断：鉴权失败，请看 JSON 中 tokenError / configDetail')
        } else if ((d.provider === 'volc-ark' || d.authMode === 'ark-bearer') && d.chatErrorMsg) {
          ElMessage.warning(`诊断：方舟调用失败 — ${d.chatErrorMsg}`)
        } else if (d.authMode === 'v2-bearer' && d.chatErrorMsg) {
          ElMessage.warning(`诊断：千帆 V2 调用失败 — ${d.chatErrorMsg}`)
        } else {
          ElMessage.warning(`诊断：对话失败 error_code=${d.chatErrorCode || '?'}`)
        }
      } else {
        ElMessage.error(res.msg || '诊断请求失败（若无权限请在「API管理」勾选诊断接口或执行内容获客同步）')
      }
    } finally {
      diagnoseLoading.value = false
    }
  }

  const copyDiagnose = async () => {
    try {
      await navigator.clipboard.writeText(diagnoseText.value)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动选择文本复制')
    }
  }

  const copyContent = async () => {
    try {
      await navigator.clipboard.writeText(syncForm.content || '')
      ElMessage.success('已复制到剪贴板')
    } catch {
      ElMessage.warning('复制失败，请手动全选复制')
    }
  }

  const goCreateArticle = () => {
    const title = String(syncForm.title || '').trim()
    const content = String(syncForm.content || '').trim()
    if (!title || !content) {
      ElMessage.warning('请先生成正文，并确认标题与正文不为空')
      return
    }
    const urls = Array.isArray(galleryUrls.value) ? galleryUrls.value.slice(0, 6) : []
    sessionStorage.setItem(
      AI_PREFILL_KEY,
      JSON.stringify({
        title: title.slice(0, 200),
        slug: String(syncForm.slug || '').trim().slice(0, 200),
        summary: String(syncForm.summary || '').slice(0, 500),
        content,
        seoTitle: String(syncForm.seoTitle || title).slice(0, 200),
        seoKeywords: String(syncForm.seoKeywords || '').slice(0, 500),
        seoDescription: String(syncForm.seoDescription || '').slice(0, 500),
        categoryId: Number(syncForm.categoryId) > 0 ? Number(syncForm.categoryId) : 0,
        coverImage: urls.join(',')
      })
    )
    router.push({ name: 'contentArticle' }).catch(() => {
      ElMessage.warning('跳转失败，请从左侧菜单打开「文章管理」')
    })
    ElMessage.success('已写入待保存数据，正在打开文章管理…')
  }
</script>

<style scoped>
  .ai-article-page {
    padding: 0;
  }
  .mb-4 {
    margin-bottom: 16px;
  }
  .max-w-3xl {
    max-width: 880px;
  }
  .sync-form {
    margin-top: 8px;
  }
  .upload-tip {
    margin-left: 10px;
    font-size: 12px;
    color: #909399;
    line-height: 1.5;
  }
  .file-meta {
    margin: 8px 0 0;
    font-size: 13px;
    color: #606266;
  }
  .field-tip {
    margin: 8px 0 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.5;
  }
  .field-tip code {
    font-size: 12px;
    background: #f4f4f5;
    padding: 1px 6px;
    border-radius: 4px;
  }
  .diagnose-pre {
    margin: 0;
    max-height: 52vh;
    overflow: auto;
    font-size: 12px;
    line-height: 1.45;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
