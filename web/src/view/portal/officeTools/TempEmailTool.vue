<template>
  <div class="tool-panel">
    <el-alert
      title="通过本站后端代理 mail.tm 临时邮箱（服务端可用）；邮件由第三方提供，仅供测试，勿用于敏感账号。"
      type="info"
      show-icon
      :closable="false"
      class="mb-3"
    />
    <div class="toolbar">
      <el-button type="primary" :loading="loadingMailbox" @click="createMailbox">申请临时邮箱</el-button>
      <el-button :loading="loadingInbox" :disabled="!mailbox" @click="refreshInbox">刷新收件箱</el-button>
      <el-button :disabled="!mailbox" @click="copyAddress">复制地址</el-button>
      <el-switch v-model="autoRefresh" active-text="自动刷新(15s)" class="ml-2" />
    </div>
    <el-descriptions v-if="mailbox" :column="1" border class="mb-3">
      <el-descriptions-item label="邮箱地址">{{ mailbox }}</el-descriptions-item>
      <el-descriptions-item label="账号 / 域名">{{ login }} @ {{ domain }}</el-descriptions-item>
    </el-descriptions>
    <el-empty v-if="!mailbox" description="点击「申请临时邮箱」获取地址" />
    <el-table v-else :data="messages" size="small" empty-text="暂无邮件">
      <el-table-column prop="from" label="发件人" min-width="160" show-overflow-tooltip />
      <el-table-column prop="subject" label="主题" min-width="200" show-overflow-tooltip />
      <el-table-column prop="date" label="时间" width="170" />
      <el-table-column label="操作" width="90">
        <template #default="{ row }">
          <el-button type="primary" link @click="readMail(row)">查看</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="dialogVisible" title="邮件内容" width="640px">
      <div v-if="mailDetail">
        <p><strong>发件人：</strong>{{ mailDetail.from }}</p>
        <p><strong>主题：</strong>{{ mailDetail.subject }}</p>
        <el-divider />
        <div class="mail-body" v-html="mailDetail.htmlBody || mailDetail.bodyHtml || mailDetail.textBody" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
  import { onUnmounted, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { createTempMailbox, getTempEmailMessages, readTempEmailMessage } from '@/api/publicPortalOffice'

  const loadingMailbox = ref(false)
  const loadingInbox = ref(false)
  const mailbox = ref('')
  const login = ref('')
  const domain = ref('')
  const token = ref('')
  const provider = ref('')
  const messages = ref([])
  const dialogVisible = ref(false)
  const mailDetail = ref(null)
  const autoRefresh = ref(false)
  let timer = null

  const createMailbox = async () => {
    loadingMailbox.value = true
    try {
      const res = await createTempMailbox()
      const data = res.data || {}
      mailbox.value = data.mailbox || ''
      login.value = data.login || ''
      domain.value = data.domain || ''
      token.value = data.token || ''
      provider.value = data.provider || ''
      messages.value = []
      if (!mailbox.value) throw new Error('未获取到邮箱')
      ElMessage.success('临时邮箱已生成')
      await refreshInbox()
    } catch (e) {
      ElMessage.error(e?.message || e?.msg || '申请失败')
    } finally {
      loadingMailbox.value = false
    }
  }

  const refreshInbox = async () => {
    if (!token.value && (!login.value || !domain.value)) return
    loadingInbox.value = true
    try {
      const res = await getTempEmailMessages({
        token: token.value || undefined,
        login: login.value,
        domain: domain.value
      })
      messages.value = res.data || []
    } catch (e) {
      ElMessage.error(e?.message || '拉取收件箱失败')
    } finally {
      loadingInbox.value = false
    }
  }

  const readMail = async (row) => {
    try {
      const res = await readTempEmailMessage({
        token: token.value || undefined,
        login: login.value,
        domain: domain.value,
        id: String(row.id)
      })
      mailDetail.value = res.data || null
      dialogVisible.value = true
    } catch {
      ElMessage.error('读取邮件失败')
    }
  }

  const copyAddress = async () => {
    if (!mailbox.value) return
    await navigator.clipboard.writeText(mailbox.value)
    ElMessage.success('已复制邮箱地址')
  }

  watch(autoRefresh, (on) => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    if (on && mailbox.value) {
      timer = setInterval(() => refreshInbox(), 15000)
    }
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; margin-bottom: 12px; }
  .mb-3 { margin-bottom: 12px; }
  .ml-2 { margin-left: 8px; }
  .mail-body { max-height: 360px; overflow: auto; line-height: 1.6; word-break: break-word; }
</style>
