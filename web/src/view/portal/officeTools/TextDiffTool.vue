<template>
  <div class="tool-panel">
    <el-row :gutter="8">
      <el-col :span="12">
        <div class="lbl">文本 A</div>
        <el-input v-model="textA" type="textarea" :rows="10" />
      </el-col>
      <el-col :span="12">
        <div class="lbl">文本 B</div>
        <el-input v-model="textB" type="textarea" :rows="10" />
      </el-col>
    </el-row>
    <el-button type="primary" size="small" class="mt-2" @click="compare">对比</el-button>
    <pre class="diff-out">{{ diffResult }}</pre>
  </div>
</template>

<script setup>
  import { ref } from 'vue'

  const textA = ref('')
  const textB = ref('')
  const diffResult = ref('')

  const compare = () => {
    const a = textA.value.split('\n')
    const b = textB.value.split('\n')
    const max = Math.max(a.length, b.length)
    const lines = []
    for (let i = 0; i < max; i++) {
      const la = a[i] ?? ''
      const lb = b[i] ?? ''
      if (la === lb) lines.push(`  ${la}`)
      else lines.push(`- ${la}\n+ ${lb}`)
    }
    diffResult.value = lines.join('\n') || '（无差异或为空）'
  }
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .lbl { font-size: 12px; color: #64748b; margin-bottom: 4px; }
  .mt-2 { margin-top: 8px; }
  .diff-out {
    margin-top: 8px;
    padding: 10px;
    background: #1e293b;
    color: #e2e8f0;
    border-radius: 6px;
    font-size: 12px;
    line-height: 1.5;
    max-height: 280px;
    overflow: auto;
    white-space: pre-wrap;
  }
</style>
