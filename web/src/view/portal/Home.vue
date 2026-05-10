<template>
  <div class="home">
    <PortalArticleFeed
      title="最新文章"
      subtitle="以下为管理端「内容 → 文章」中已发布内容，点击卡片阅读全文。无需登录。"
      show-search
      show-pager
      :page-size="10"
    />
  </div>

   <div class="home">
    <PortalArticleHot
      title="热门文章"
      subtitle="以下为管理端「内容 → 文章」中已发布内容，点击卡片阅读全文。无需登录。"
 
      show-pager
      :page-size="10"
    />
  </div>
</template>

<script setup>
  import PortalArticleFeed from '@/view/portal/components/PortalArticleFeed.vue'
  import PortalArticleHot from '@/view/portal/components/PortalArticleHot.vue'
</script>

<style scoped>
  .home {
    min-height: 40vh;
  }
</style>
