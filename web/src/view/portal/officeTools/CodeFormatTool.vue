<template>
  <div class="tool-panel">
    <el-tabs v-model="tab">
      <el-tab-pane label="JavaScript" name="js">
        <div class="btns">
          <el-button type="primary" size="small" @click="fmtJs">格式化</el-button>
          <el-button size="small" @click="minJs">压缩</el-button>
        </div>
        <el-row :gutter="8" class="mt-2">
          <el-col :span="12"><el-input v-model="jsIn" type="textarea" :rows="12" /></el-col>
          <el-col :span="12"><el-input v-model="jsOut" type="textarea" :rows="12" readonly /></el-col>
        </el-row>
      </el-tab-pane>
      <el-tab-pane label="CSS" name="css">
        <div class="btns">
          <el-button type="primary" size="small" @click="cssOut = formatCSS(cssIn)">格式化</el-button>
          <el-button size="small" @click="cssOut = minifyCSS(cssIn)">压缩</el-button>
        </div>
        <el-row :gutter="8" class="mt-2">
          <el-col :span="12"><el-input v-model="cssIn" type="textarea" :rows="12" /></el-col>
          <el-col :span="12"><el-input v-model="cssOut" type="textarea" :rows="12" readonly /></el-col>
        </el-row>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { formatCSS, formatJS, minifyCSS, minifyJS } from '@/utils/officeCodec'

  const tab = ref('js')
  const jsIn = ref('function hello(){console.log("hi");}')
  const jsOut = ref('')
  const cssIn = ref('.box { color: red; margin: 0; }')
  const cssOut = ref('')

  const fmtJs = () => {
    try {
      jsOut.value = formatJS(jsIn.value)
    } catch {
      jsOut.value = jsIn.value
    }
  }
  const minJs = () => {
    jsOut.value = minifyJS(jsIn.value)
  }
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .btns { display: flex; gap: 8px; }
  .mt-2 { margin-top: 8px; }
</style>
