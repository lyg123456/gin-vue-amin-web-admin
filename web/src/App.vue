<template>
  <div id="app" class="text-slate-700 !dark:text-slate-500">
    <el-config-provider :locale="zhCn" :size="appStore.config.global_size">
      <router-view />
      <Application />
    </el-config-provider>
  </div>
</template>

<script setup>
  import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
  import Application from '@/components/application/index.vue'
  import { useAppStore } from '@/pinia'

  const appStore = useAppStore()
  defineOptions({
    name: 'App'
  })
</script>
<style lang="scss">
  // 引入初始化样式
  body {
    background-color: #f5f5f5;
  }

  html.dark body {
    background-color: rgb(30 41 59);
  }

  /* min-height + 不锁死 overflow，门户 /web 可整页滚动；管理端内部仍有各自 overflow-auto */
  #app {
    min-height: 100vh;
    overflow-x: hidden;
    font-weight: 400 !important;
    background-color: #f5f5f5;
  }

  html.dark #app {
    background-color: rgb(30 41 59);
  }

  .el-button {
    font-weight: 400 !important;
  }

  .gva-body-h {
    min-height: calc(100% - 3rem);
  }

  .gva-container {
    height: calc(100% - 2.5rem);
  }

  .gva-container2 {
    height: calc(100% - 4.5rem);
  }
</style>
