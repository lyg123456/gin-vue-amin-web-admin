<template>
  <div class="gva-table-box">
    <el-alert
      title="异步成片任务：API 入队 → Redis 列表 → 分发协程 BLPOP → channel → Worker 池执行 DashScope 成片。"
      type="info"
      show-icon
      :closable="false"
      class="mb-4"
    />
    <div class="gva-btn-list flex gap-2 flex-wrap items-center mb-3">
      <el-tag v-if="asyncEnabled" type="success">异步已启用</el-tag>
      <el-tag v-else type="warning">异步未启用</el-tag>
      <el-tag type="info">Redis 队列长度：{{ redisQueueLen }}</el-tag>
      <el-select v-model="search.status" clearable placeholder="任务状态" style="width: 140px" @change="load">
        <el-option label="排队中" value="queued" />
        <el-option label="处理中" value="processing" />
        <el-option label="成功" value="succeeded" />
        <el-option label="失败" value="failed" />
      </el-select>
      <el-input
        v-model="search.shortVideoId"
        clearable
        placeholder="短视频 ID"
        style="width: 120px"
        @keyup.enter="load"
      />
      <el-button type="primary" @click="load">刷新</el-button>
      <el-button @click="$router.push({ name: 'shortVideoList' })">返回短视频列表</el-button>
    </div>

    <el-table :data="tableData" row-key="ID">
      <el-table-column label="任务ID" prop="ID" width="80" />
      <el-table-column label="短视频ID" prop="shortVideoId" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="jobStatusType(row.status)" size="small">{{ jobStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="提供商" prop="provider" width="100" />
      <el-table-column label="第三方任务ID" prop="externalTaskId" min-width="140" show-overflow-tooltip />
      <el-table-column label="尝试次数" prop="attempts" width="90" />
      <el-table-column label="失败原因" prop="errorMsg" min-width="160" show-overflow-tooltip />
      <el-table-column label="入队时间" width="170">
        <template #default="{ row }">{{ formatTime(row.enqueuedAt || row.CreatedAt) }}</template>
      </el-table-column>
      <el-table-column label="完成时间" width="170">
        <template #default="{ row }">{{ formatTime(row.finishedAt) }}</template>
      </el-table-column>
    </el-table>

    <div class="gva-pagination">
      <el-pagination
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="onPage"
      />
    </div>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { getVideoGenJobList } from '@/api/contentVideoGenJob'

  defineOptions({ name: 'ShortVideoGenJobs' })

  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const tableData = ref([])
  const redisQueueLen = ref(0)
  const asyncEnabled = ref(false)
  const search = ref({ status: '', shortVideoId: '' })

  const jobStatusLabel = (s) =>
    ({ queued: '排队', processing: '处理中', succeeded: '成功', failed: '失败' })[s] || s
  const jobStatusType = (s) =>
    ({ queued: 'info', processing: 'warning', succeeded: 'success', failed: 'danger' })[s] || 'info'

  const formatTime = (t) => {
    if (!t) return '--'
    const d = new Date(t)
    return Number.isNaN(d.getTime()) ? '--' : d.toLocaleString('zh-CN')
  }

  const load = async () => {
    const res = await getVideoGenJobList({
      page: page.value,
      pageSize: pageSize.value,
      status: search.value.status || undefined,
      shortVideoId: search.value.shortVideoId || undefined
    })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      redisQueueLen.value = res.data.redisQueueLen ?? 0
      asyncEnabled.value = !!res.data.asyncEnabled
    }
  }

  const onPage = (p) => {
    page.value = p
    load()
  }

  onMounted(load)
</script>

<style scoped>
  .mb-4 { margin-bottom: 16px; }
  .mb-3 { margin-bottom: 12px; }
</style>
