<template>
  <div class="tool-panel">
    <el-tabs v-model="tab">
      <el-tab-pane label="Base64" name="b64">
        <el-input v-model="text" type="textarea" :rows="5" placeholder="输入文本" />
        <div class="btns">
          <el-button type="primary" size="small" @click="doB64Enc">编码</el-button>
          <el-button size="small" @click="doB64Dec">解码</el-button>
          <el-button size="small" @click="copy(out)">复制</el-button>
        </div>
        <el-input v-model="out" type="textarea" :rows="5" readonly />
      </el-tab-pane>
      <el-tab-pane label="URL" name="url">
        <el-input v-model="text" type="textarea" :rows="5" />
        <div class="btns">
          <el-button type="primary" size="small" @click="out = encodeURIComponent(text)">编码</el-button>
          <el-button size="small" @click="out = decodeURIComponent(text.replace(/\+/g, ' '))">解码</el-button>
        </div>
        <el-input v-model="out" type="textarea" :rows="5" readonly />
      </el-tab-pane>
      <el-tab-pane label="MD5" name="md5">
        <el-input v-model="text" type="textarea" :rows="6" />
        <el-button type="primary" size="small" class="mt-2" @click="out = md5Hex(text)">计算 MD5</el-button>
        <el-input v-model="out" class="mt-2" readonly />
      </el-tab-pane>
      <el-tab-pane label="Unicode" name="uni">
        <el-input v-model="text" type="textarea" :rows="5" />
        <div class="btns">
          <el-button size="small" @click="out = chineseToUnicode(text)">中文→Unicode</el-button>
          <el-button size="small" @click="out = unicodeToChinese(text)">Unicode→中文</el-button>
        </div>
        <el-input v-model="out" type="textarea" :rows="5" readonly />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { base64Decode, base64Encode, chineseToUnicode, md5Hex, unicodeToChinese } from '@/utils/officeCodec'

  const tab = ref('b64')
  const text = ref('')
  const out = ref('')

  const doB64Enc = () => {
    try {
      out.value = base64Encode(text.value)
    } catch {
      ElMessage.error('编码失败')
    }
  }
  const doB64Dec = () => {
    try {
      out.value = base64Decode(text.value)
    } catch {
      ElMessage.error('解码失败')
    }
  }
  const copy = async (t) => {
    await navigator.clipboard.writeText(t)
    ElMessage.success('已复制')
  }
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .btns { margin: 8px 0; display: flex; gap: 8px; flex-wrap: wrap; }
  .mt-2 { margin-top: 8px; }
</style>
