<template>
  <div class="colorblind-page">
    <section class="portal-panel">
      <div class="page-title-bar">
        <h1 class="page-h1">色盲色弱测试</h1>
        <p class="page-desc">采用 Wikimedia 公开石原氏检测图 · 结果仅供参考，不能替代医学检查</p>
      </div>

      <div class="panel-body">
        <el-tabs v-model="tab">
          <el-tab-pane label="自测筛查" name="test">
            <p class="hint">
              请观察下方<strong>标准石原氏色盲图</strong>中的数字，选择您看到的内容（共 {{ plates.length }} 题）。
              参考答案按<strong>正常色觉</strong>标注；色盲/色弱者可能看到不同数字，以医院检查为准。
            </p>
            <div v-for="(p, idx) in plates" :key="p.id" class="plate-card">
              <div class="plate-title">{{ p.title }}</div>
              <div class="plate-img-wrap">
                <img
                  :src="imgSrc(p)"
                  :alt="p.title"
                  class="plate-img"
                  loading="lazy"
                  referrerpolicy="no-referrer"
                  @error="(e) => onImgError(p, e)"
                />
                <p v-if="imgErrors[p.id]" class="img-fail">图片加载失败，请检查网络或运行下载脚本后刷新</p>
              </div>
              <el-radio-group v-model="answers[idx]" class="plate-options">
                <el-radio v-for="opt in p.options" :key="opt" :label="opt">{{ opt }}</el-radio>
              </el-radio-group>
            </div>
            <el-button type="primary" @click="submitTest">查看自测结果</el-button>
            <el-alert v-if="testResult" :title="testResult" :type="testType" show-icon class="mt-3" />
            <p class="source-note">
              图源：
              <a href="https://commons.wikimedia.org/wiki/Category:Ishihara_plates" target="_blank" rel="noopener noreferrer">Wikimedia Commons — Ishihara plates</a>
            </p>
          </el-tab-pane>

          <el-tab-pane label="色觉模拟" name="sim">
            <p class="hint">选择色觉类型，查看常见颜色在不同色觉下的近似效果（RGB 变换矩阵）。</p>
            <el-radio-group v-model="simType" class="mb-3">
              <el-radio-button label="normal">正常</el-radio-button>
              <el-radio-button label="protanopia">红色盲</el-radio-button>
              <el-radio-button label="deuteranopia">绿色盲</el-radio-button>
              <el-radio-button label="tritanopia">蓝色盲</el-radio-button>
              <el-radio-button label="protanomaly">红色弱</el-radio-button>
              <el-radio-button label="deuteranomaly">绿色弱</el-radio-button>
            </el-radio-group>
            <div class="color-grid">
              <div
                v-for="c in sampleColors"
                :key="c.name"
                class="color-chip"
                :style="{ background: transformColor(c.hex) }"
              >
                <span>{{ c.name }}</span>
              </div>
            </div>
            <div class="upload-sim mt-3">
              <el-upload :auto-upload="false" :show-file-list="false" accept="image/*" @change="onImagePick">
                <el-button>上传图片预览模拟效果</el-button>
              </el-upload>
              <div v-if="previewUrl" class="sim-images">
                <div class="sim-col">
                  <span>原图</span>
                  <img :src="previewUrl" alt="原图" />
                </div>
                <div class="sim-col">
                  <span>{{ simLabel }}</span>
                  <img :src="previewUrl" alt="模拟" :style="{ filter: simFilter }" />
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="类型说明" name="info">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="红色盲 (Protanopia)">难以分辨红绿，红色感知弱。</el-descriptions-item>
              <el-descriptions-item label="绿色盲 (Deuteranopia)">红绿混淆，最常见。</el-descriptions-item>
              <el-descriptions-item label="蓝色盲 (Tritanopia)">蓝黄难辨，较少见。</el-descriptions-item>
              <el-descriptions-item label="建议">自测异常请到医院眼科做正规 Ishihara 检查。</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
        </el-tabs>
      </div>
    </section>

    <svg class="svg-filters" aria-hidden="true">
      <defs>
        <filter id="cb-protanopia">
          <feColorMatrix type="matrix" values="0.567,0.433,0,0,0 0.558,0.442,0,0,0 0,0.242,0.758,0,0 0,0,0,1,0" />
        </filter>
        <filter id="cb-deuteranopia">
          <feColorMatrix type="matrix" values="0.625,0.375,0,0,0 0.7,0.3,0,0,0 0,0.3,0.7,0,0 0,0,0,1,0" />
        </filter>
        <filter id="cb-tritanopia">
          <feColorMatrix type="matrix" values="0.95,0.05,0,0,0 0,0.433,0.567,0,0 0,0.475,0.525,0,0 0,0,0,1,0" />
        </filter>
        <filter id="cb-protanomaly">
          <feColorMatrix type="matrix" values="0.817,0.183,0,0,0 0.333,0.667,0,0,0 0,0.125,0.875,0,0 0,0,0,1,0" />
        </filter>
        <filter id="cb-deuteranomaly">
          <feColorMatrix type="matrix" values="0.8,0.2,0,0,0 0.258,0.742,0,0,0 0,0.142,0.858,0,0 0,0,0,1,0" />
        </filter>
      </defs>
    </svg>
  </div>
</template>

<script setup>
  import { computed, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { ishiharaPlates, plateRemoteUrl } from '@/constants/ishiharaPlates'

  defineOptions({ name: 'WebColorBlindness' })

  const tab = ref('test')
  const simType = ref('deuteranopia')
  const previewUrl = ref('')
  const answers = ref(ishiharaPlates.map(() => ''))
  const testResult = ref('')
  const testType = ref('info')
  const imgErrors = reactive({})
  /** 优先本地（download-ishihara.ps1），否则 Wikimedia 在线图 */
  const useRemote = reactive({})

  const plates = ishiharaPlates

  const sampleColors = [
    { name: '红', hex: '#e53935' },
    { name: '绿', hex: '#43a047' },
    { name: '蓝', hex: '#1e88e5' },
    { name: '黄', hex: '#fdd835' },
    { name: '橙', hex: '#fb8c00' },
    { name: '紫', hex: '#8e24aa' },
    { name: '青', hex: '#00acc1' },
    { name: '粉', hex: '#ec407a' }
  ]

  const matrices = {
    normal: null,
    protanopia: [0.567, 0.433, 0, 0.558, 0.442, 0, 0, 0.242, 0.758],
    deuteranopia: [0.625, 0.375, 0, 0.7, 0.3, 0, 0, 0.3, 0.7],
    tritanopia: [0.95, 0.05, 0, 0, 0.433, 0.567, 0, 0.475, 0.525],
    protanomaly: [0.817, 0.183, 0, 0.333, 0.667, 0, 0, 0.125, 0.875],
    deuteranomaly: [0.8, 0.2, 0, 0.258, 0.742, 0, 0, 0.142, 0.858]
  }

  const simLabels = {
    normal: '正常视觉',
    protanopia: '红色盲模拟',
    deuteranopia: '绿色盲模拟',
    tritanopia: '蓝色盲模拟',
    protanomaly: '红色弱模拟',
    deuteranomaly: '绿色弱模拟'
  }

  const simLabel = computed(() => simLabels[simType.value] || '')

  const simFilter = computed(() => {
    if (simType.value === 'normal') return 'none'
    return `url(#cb-${simType.value})`
  })

  const imgSrc = (p) => {
    if (useRemote[p.id]) return plateRemoteUrl(p.file, 420)
    return p.local
  }

  const onImgError = (p, e) => {
    if (!useRemote[p.id]) {
      useRemote[p.id] = true
      e.target.src = plateRemoteUrl(p.file, 420)
      return
    }
    imgErrors[p.id] = true
  }

  const hexToRgb = (hex) => {
    const h = hex.replace('#', '')
    return [parseInt(h.slice(0, 2), 16), parseInt(h.slice(2, 4), 16), parseInt(h.slice(4, 6), 16)]
  }

  const rgbToHex = (r, g, b) => {
    const f = (n) => Math.max(0, Math.min(255, Math.round(n))).toString(16).padStart(2, '0')
    return `#${f(r)}${f(g)}${f(b)}`
  }

  const transformColor = (hex) => {
    const m = matrices[simType.value]
    if (!m) return hex
    const [r, g, b] = hexToRgb(hex)
    return rgbToHex(r * m[0] + g * m[1] + b * m[2], r * m[3] + g * m[4] + b * m[5], r * m[6] + g * m[7] + b * m[8])
  }

  const submitTest = () => {
    if (answers.value.some((a) => !a)) {
      ElMessage.warning('请完成全部题目')
      return
    }
    let wrong = 0
    plates.forEach((p, i) => {
      if (answers.value[i] !== p.answer) wrong++
    })
    if (wrong === 0) {
      testResult.value = '自测结果：各题辨认与参考一致，色觉大致正常（仍建议定期眼科检查）。'
      testType.value = 'success'
    } else if (wrong >= 3) {
      testResult.value = `自测结果：${wrong} 题与参考不一致，可能存在色盲或色弱，建议到医院眼科检查。`
      testType.value = 'warning'
    } else {
      testResult.value = `自测结果：${wrong} 题不一致，可能有轻度色觉异常，建议进一步检查。`
      testType.value = 'info'
    }
  }

  const onImagePick = (f) => {
    const raw = f?.raw
    if (!raw) return
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = URL.createObjectURL(raw)
  }
</script>

<style lang="scss" scoped>
  .colorblind-page {
    padding-bottom: 16px;
  }
  .portal-panel {
    background: var(--portal-panel-bg, #fff);
    border-radius: var(--portal-radius, 4px);
    border: 1px solid #e0e0e0;
    padding: 0;
    overflow: hidden;
  }
  .page-title-bar {
    padding: 14px 16px 12px;
    border-bottom: 1px solid #eee;
    background: #fafafa;
    text-align: center;
  }
  .page-h1 {
    margin: 0;
    font-size: var(--portal-font-title, 18px);
    font-weight: 700;
    color: #333;
  }
  .page-desc {
    margin: 6px 0 0;
    font-size: 14px;
    color: #888;
  }
  .panel-body {
    padding: 16px;
  }
  .hint {
    font-size: 14px;
    color: #666;
    margin: 0 0 14px;
    line-height: 1.6;
    code {
      font-size: 12px;
      background: #f0f0f0;
      padding: 2px 6px;
      border-radius: 3px;
    }
  }
  .plate-card {
    margin-bottom: 24px;
    padding: 16px;
    border: 1px solid #e8e8e8;
    border-radius: 4px;
    background: #fafafa;
    text-align: center;
  }
  .plate-title {
    font-size: 15px;
    font-weight: 600;
    margin-bottom: 12px;
    color: #333;
  }
  .plate-img-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 14px;
  }
  .plate-img {
    width: 280px;
    height: 280px;
    max-width: 100%;
    object-fit: contain;
    border-radius: 50%;
    border: 1px solid #ddd;
    background: #fff;
  }
  .img-fail {
    margin: 8px 0 0;
    font-size: 13px;
    color: #e6a23c;
  }
  .plate-options {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px 20px;
  }
  .source-note {
    margin-top: 16px;
    font-size: 13px;
    color: #888;
    a {
      color: var(--portal-brand, #1a73e8);
    }
  }
  .color-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  .color-chip {
    width: 88px;
    height: 56px;
    border-radius: 4px;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding-bottom: 6px;
    border: 1px solid rgba(0, 0, 0, 0.08);
    span {
      font-size: 13px;
      color: #fff;
      text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
    }
  }
  .sim-images {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    margin-top: 14px;
  }
  .sim-col {
    flex: 1;
    min-width: 200px;
    text-align: center;
    span {
      display: block;
      font-size: 13px;
      color: #666;
      margin-bottom: 8px;
    }
    img {
      max-width: 100%;
      border: 1px solid #e0e0e0;
      border-radius: 4px;
    }
  }
  .svg-filters {
    position: absolute;
    width: 0;
    height: 0;
    overflow: hidden;
  }
  .mt-3 {
    margin-top: 14px;
  }
  .mb-3 {
    margin-bottom: 12px;
  }
</style>
