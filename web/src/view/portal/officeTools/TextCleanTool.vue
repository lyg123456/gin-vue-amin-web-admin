<template>
  <div class="tool-panel">
    <el-input v-model="text" type="textarea" :rows="8" />
    <div class="btns">
      <el-button size="small" @click="removeEmpty">去除空行</el-button>
      <el-button size="small" @click="dedupe">文本去重</el-button>
      <el-button size="small" @click="trimLines">行首尾去空格</el-button>
      <el-button size="small" @click="oneLine">多行合一行</el-button>
    </div>
  </div>
</template>

<script setup>
  import { ref } from 'vue'

  const text = ref('')

  const removeEmpty = () => {
    text.value = text.value
      .split('\n')
      .filter((l) => l.trim() !== '')
      .join('\n')
  }
  const dedupe = () => {
    const seen = new Set()
    text.value = text.value
      .split('\n')
      .filter((l) => {
        const k = l.trim()
        if (!k || seen.has(k)) return false
        seen.add(k)
        return true
      })
      .join('\n')
  }
  const trimLines = () => {
    text.value = text.value.split('\n').map((l) => l.trim()).join('\n')
  }
  const oneLine = () => {
    text.value = text.value.replace(/\s*\n\s*/g, ' ').trim()
  }
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .btns { margin-top: 8px; display: flex; flex-wrap: wrap; gap: 8px; }
</style>
