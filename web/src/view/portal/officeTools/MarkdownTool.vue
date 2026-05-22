<template>
  <div class="tool-panel">
    <el-row :gutter="12">
      <el-col :xs="24" :md="12">
        <el-input v-model="md" type="textarea" :rows="14" placeholder="Markdown 源码" />
      </el-col>
      <el-col :xs="24" :md="12">
        <div class="preview md-body" v-html="html" />
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { marked } from 'marked'

  const md = ref('# 标题\n\n**dnf1688** 在线工具\n\n- 列表项\n- [链接](https://example.com)')

  const html = computed(() => {
    try {
      return marked.parse(md.value || '')
    } catch {
      return '<p>解析失败</p>'
    }
  })
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .preview {
    min-height: 280px;
    padding: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    background: #fafafa;
    font-size: 14px;
    line-height: 1.6;
  }
  .md-body :deep(pre) {
    background: #1e293b;
    color: #f8fafc;
    padding: 8px;
    border-radius: 4px;
    overflow: auto;
  }
</style>
