<template>
  <div class="tool-panel">
    <el-tabs v-model="tab">
      <el-tab-pane label="UUID" name="uuid">
        <el-button type="primary" size="small" @click="addUuid">生成 UUID</el-button>
        <el-button size="small" @click="copyList">复制全部</el-button>
        <el-input v-model="uuidList" type="textarea" :rows="8" class="mt-2" readonly />
      </el-tab-pane>
      <el-tab-pane label="随机密码" name="pwd">
        <el-row :gutter="8">
          <el-col :span="8"><el-input-number v-model="pwdLen" :min="6" :max="64" class="w-full" /></el-col>
          <el-col :span="16">
            <el-checkbox v-model="pwdSymbol">含符号</el-checkbox>
          </el-col>
        </el-row>
        <el-button type="primary" size="small" class="mt-2" @click="genPwd">生成</el-button>
        <el-input v-model="pwdOut" class="mt-2" readonly />
      </el-tab-pane>
      <el-tab-pane label="驼峰/下划线" name="case">
        <el-input v-model="caseIn" type="textarea" :rows="4" placeholder="user_name 或 userName" />
        <div class="btns">
          <el-button size="small" @click="caseOut = toCamel(caseIn)">转驼峰</el-button>
          <el-button size="small" @click="caseOut = toSnake(caseIn)">转下划线</el-button>
          <el-button size="small" @click="caseOut = toKebab(caseIn)">转短横线</el-button>
        </div>
        <el-input v-model="caseOut" readonly />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { genPassword, genUuid, toCamel, toKebab, toSnake } from '@/utils/officeCodec'

  const tab = ref('uuid')
  const uuidList = ref('')
  const pwdLen = ref(16)
  const pwdSymbol = ref(true)
  const pwdOut = ref('')
  const caseIn = ref('user_name_example')
  const caseOut = ref('')

  const addUuid = () => {
    uuidList.value = (uuidList.value ? uuidList.value + '\n' : '') + genUuid()
  }
  const copyList = async () => {
    await navigator.clipboard.writeText(uuidList.value)
    ElMessage.success('已复制')
  }
  const genPwd = () => {
    pwdOut.value = genPassword(pwdLen.value, { symbol: pwdSymbol.value })
  }
</script>

<style scoped>
  .tool-panel { padding: 4px 0; }
  .btns { margin: 8px 0; display: flex; gap: 8px; flex-wrap: wrap; }
  .mt-2 { margin-top: 8px; }
  .w-full { width: 100%; }
</style>
