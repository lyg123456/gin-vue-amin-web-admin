<template>
  <el-alert type="warning" show-icon :closable="false" class="crawler-disclaimer">
    <template #title>数据使用声明</template>
    {{ PORTAL_CRAWLER_DATA_NOTICE }}
  </el-alert>
</template>

<script setup>
  import { PORTAL_CRAWLER_DATA_NOTICE } from '@/constants/portalDisclaimer'

  defineOptions({ name: 'CrawlerDisclaimerAlert' })
</script>

<style scoped>
  .crawler-disclaimer {
    margin-bottom: 12px;
  }
</style>
