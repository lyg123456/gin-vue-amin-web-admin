<template>
  <div class="member-center">
    <section class="panel article-panel">
      <PortalArticleFeed
        title="获客内容 · 站点文章"
        subtitle="与首页同源：展示管理端「内容 → 文章」中已发布的文章，访客无需登录即可浏览。"
        :show-search="false"
        show-pager
        :page-size="8"
      />
      <div class="more-row">
        <router-link class="more-link" to="/">前往首页 · 搜索与浏览全部</router-link>
      </div>
    </section>
    
     
<!-- 
    <section class="panel">
      <h1>会员中心</h1>
      <p class="hint">浏览文章无需登录；登录后可使用后续会员功能（收藏、评论等可对接扩展）。</p>

      <div v-if="token" class="user-card">
        <img v-if="avatar" :src="avatar" alt="" class="avatar" />
        <div v-else class="avatar placeholder">{{ initials }}</div>
        <div class="info">
          <div class="name">{{ displayName }}</div>
          <div class="id-line" v-if="userInfo.ID">会员 ID：{{ userInfo.ID }}</div>
        </div>
        <div class="actions">
          <el-button type="danger" plain @click="onLogout">退出登录</el-button>
        </div>
      </div>

      <div v-else class="guest">
        <p>您还未登录会员账号。</p>
        <div class="btn-row">
          <el-button type="primary" @click="router.push({ name: 'MemberEntryLogin' })">会员登录</el-button>
          <el-button @click="router.push({ name: 'WebMemberRegister' })">注册说明</el-button>
        </div>
      </div>
    </section> -->

    
  </div>
</template>

<script setup>
  import { computed, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/pinia/modules/user'
  import { getUrl } from '@/utils/image'
  import PortalArticleFeed from '@/view/portal/components/PortalArticleFeed.vue'

  const router = useRouter()
  const userStore = useUserStore()
  const { token, userInfo } = storeToRefs(userStore)

  const displayName = computed(() => {
    const n = userInfo.value?.nickName
    if (n) return n
    if (userInfo.value?.userName) return userInfo.value.userName
    return '会员'
  })

  const avatar = computed(() => {
    const h = userInfo.value?.headerImg
    return h ? getUrl(h) : ''
  })

  const initials = computed(() => {
    const n = displayName.value
    return n ? n.slice(0, 1).toUpperCase() : '?'
  })

  const onLogout = () => {
    ElMessageBox.confirm('确定退出当前会员登录？', '提示', {
      type: 'warning',
      confirmButtonText: '退出',
      cancelButtonText: '取消'
    }).then(async () => {
      await userStore.MemberLogout()
    })
  }

  onMounted(async () => {
    if (token.value) {
      await userStore.GetUserInfo()
    }
  })
</script>

<style scoped>
  .member-center {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  .panel {
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    border: none;
    padding: 24px;
  }
  .muted-panel {
    background: var(--portal-muted-bg, #ffffff);
  }
  h1 {
    margin: 0 0 8px;
    font-size: 1.5rem;
  }
  h2 {
    margin: 0 0 12px;
    font-size: 1.1rem;
  }
  .hint {
    margin: 0 0 20px;
    color: var(--portal-text-secondary, #6b7280);
    font-size: 0.9rem;
  }
  .user-card {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 16px;
  }
  .avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    object-fit: cover;
  }
  .avatar.placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background: #e0e7ff;
    color: #3730a3;
    font-weight: 700;
    font-size: 1.25rem;
  }
  .info {
    flex: 1;
    min-width: 160px;
  }
  .name {
    font-weight: 600;
    font-size: 1.1rem;
  }
  .id-line {
    font-size: 0.85rem;
    color: var(--portal-text-meta, #9ca3af);
    margin-top: 4px;
  }
  .actions {
    width: 100%;
    flex-basis: 100%;
    padding-top: 8px;
    border-top: 1px solid var(--portal-hairline, #f3f4f6);
    margin-top: 4px;
  }
  .guest p {
    margin: 0 0 16px;
    color: var(--portal-text-body, #4b5563);
    line-height: 1.6;
  }
  .btn-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  ul {
    margin: 0;
    padding-left: 1.2rem;
    color: var(--portal-text-body, #4b5563);
    line-height: 1.8;
  }
  .article-panel {
    padding-bottom: 16px;
  }
  .more-row {
    margin-top: 8px;
    text-align: right;
  }
  .more-link {
    font-size: 0.9rem;
    color: var(--portal-link, #2563eb);
    text-decoration: none;
  }
  .more-link:hover {
    text-decoration: underline;
  }
</style>
