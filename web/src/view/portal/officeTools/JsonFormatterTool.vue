<template>
  <div class="tool-panel">
    <div class="toolbar">
      <el-button type="primary" @click="formatJson">格式化</el-button>
      <el-button @click="minifyJson">压缩</el-button>
      <el-button @click="validateJson">校验</el-button>
      <el-button @click="copyOut">复制结果</el-button>
      <el-button @click="clearAll">清空</el-button>
    </div>
    <el-row :gutter="12" class="editor-row">
      <el-col :span="12">
        <div class="label">输入</div>
        <el-input v-model="input" type="textarea" :rows="14" placeholder="粘贴 JSON" />
      </el-col>
      <el-col :span="12">
        <div class="label">输出</div>
        <el-input v-model="output" type="textarea" :rows="14" readonly placeholder="格式化结果" />
      </el-col>
    </el-row>
    <el-alert v-if="message" :title="message" :type="alertType" show-icon class="mt-2" @close="message = ''" />
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'

  const input = ref('{\n  "name": "示例",\n  "items": [1, 2, 3]\n}')
  const output = ref('')
  const message = ref('')
  const alertType = ref('success')

  const parseInput = () => {
    const text = input.value.trim()
    if (!text) throw new Error('请输入 JSON')
    return JSON.parse(text)
  }

  const formatJson = () => {
    try {
      const obj = parseInput()
      output.value = JSON.stringify(obj, null, 2)
      message.value = '格式化成功'
      alertType.value = 'success'
    } catch (e) {
      message.value = e.message || 'JSON 解析失败'
      alertType.value = 'error'
    }
  }

  const minifyJson = () => {
    try {
      const obj = parseInput()
      output.value = JSON.stringify(obj)
      message.value = '已压缩为一行'
      alertType.value = 'success'
    } catch (e) {
      message.value = e.message || 'JSON 解析失败'
      alertType.value = 'error'
    }
  }

  const validateJson = () => {
    try {
      parseInput()
      message.value = 'JSON 语法正确'
      alertType.value = 'success'
    } catch (e) {
      message.value = e.message || 'JSON 无效'
      alertType.value = 'error'
    }
  }

  const copyOut = async () => {
    const text = output.value || input.value
    if (!text) {
      ElMessage.warning('无内容可复制')
      return
    }
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  }

  const clearAll = () => {
    input.value = ''
    output.value = ''
    message.value = ''
  }

  formatJson()
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .toolbar { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
  .label { font-size: 13px; color: #606266; margin-bottom: 6px; }
  .mt-2 { margin-top: 8px; }
</style>
