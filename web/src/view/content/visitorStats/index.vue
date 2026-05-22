<template>
  <div class="gva-table-box">
    <el-alert
      title="统计门户访问：用户端请求 /api/public/web/stats 时按 IP + 自然日累加访问次数。若侧边栏没有本页，请点击下方「同步菜单/权限」后重新登录。"
      type="info"
      show-icon
      :closable="false"
      class="mb-4"
    />

    <el-row :gutter="16" class="summary-row mb-2">
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic title="总访问 PV" :value="summary.totalPv">
          <template #suffix>
            <span class="stat-tip">次</span>
          </template>
        </el-statistic>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic title="总访客 UV" :value="summary.totalUv">
          <template #suffix>
            <span class="stat-tip">个 IP</span>
          </template>
        </el-statistic>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic title="今日 PV" :value="summary.todayPv" />
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic title="今日 UV" :value="summary.todayUv" />
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic :title="filterDateLabel + ' PV'" :value="summary.filterDatePv" />
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-statistic :title="filterDateLabel + ' UV'" :value="summary.filterDateUv" />
      </el-col>
    </el-row>

    <div class="gva-btn-list flex gap-2 flex-wrap items-center">
      <el-date-picker
        v-model="search.visitDate"
        type="date"
        value-format="YYYY-MM-DD"
        placeholder="访问日期"
        clearable
        style="width: 160px"
      />
      <el-input
        v-model="search.keyword"
        clearable
        placeholder="IP 关键字"
        style="width: 200px"
        @keyup.enter="getTableData"
      />
      <el-button type="primary" icon="search" @click="getTableData">查询</el-button>
      <el-button @click="resetSearch">重置</el-button>
      <el-button type="warning" :loading="syncLoading" @click="onSyncMenu">同步菜单/权限</el-button>
    </div>

    <el-table :data="tableData" row-key="ID">
      <el-table-column align="left" label="ID" min-width="70" prop="ID" />
      <el-table-column align="left" label="访客 IP" min-width="140" prop="clientIp" />
      <el-table-column align="left" label="访问日期" min-width="120" prop="visitDate" />
      <el-table-column align="left" label="当日访问次数" min-width="120" prop="visitCount" />
      <el-table-column align="left" label="最近更新" min-width="170">
        <template #default="{ row }">
          {{ formatTime(row.UpdatedAt) }}
        </template>
      </el-table-column>
    </el-table>

    <div class="gva-pagination">
      <el-pagination
        :current-page="page"
        :page-size="pageSize"
        :page-sizes="[10, 30, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handleCurrentChange"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { getPortalVisitorList } from '@/api/contentPortalVisitor'
  import { syncContentInit } from '@/api/contentInit'

  defineOptions({ name: 'ContentPortalVisitor' })

  const todayStr = () => {
    const d = new Date()
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const search = ref({ visitDate: todayStr(), keyword: '' })
  const syncLoading = ref(false)

  const summary = ref({
    totalPv: 0,
    totalUv: 0,
    todayPv: 0,
    todayUv: 0,
    filterDatePv: 0,
    filterDateUv: 0,
    visitDate: ''
  })

  const filterDateLabel = computed(() => search.value.visitDate || '筛选日')

  const formatTime = (t) => {
    if (!t) return '--'
    const d = new Date(t)
    if (Number.isNaN(d.getTime())) return '--'
    return d.toLocaleString('zh-CN')
  }

  const getTableData = async () => {
    const res = await getPortalVisitorList({
      page: page.value,
      pageSize: pageSize.value,
      visitDate: search.value.visitDate || undefined,
      keyword: search.value.keyword || undefined
    })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      page.value = res.data.page || 1
      pageSize.value = res.data.pageSize || 10
      if (res.data.summary) {
        summary.value = res.data.summary
      }
    }
  }

  const onSyncMenu = () => {
    ElMessageBox.confirm(
      '将把「内容获客」相关菜单、API、权限增量写入数据库（含访客统计）。确定执行吗？',
      '同步确认',
      { type: 'warning', confirmButtonText: '确定同步', cancelButtonText: '取消' }
    ).then(async () => {
      syncLoading.value = true
      try {
        const res = await syncContentInit()
        if (res.code === 0) {
          ElMessage.success('同步成功，请退出重新登录后查看侧边栏菜单')
        }
      } finally {
        syncLoading.value = false
      }
    })
  }

  const resetSearch = () => {
    search.value = { visitDate: todayStr(), keyword: '' }
    page.value = 1
    getTableData()
  }

  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  onMounted(getTableData)
</script>

<style scoped>
  .mb-4 { margin-bottom: 16px; }
  .summary-row :deep(.el-statistic__head) { font-size: 13px; }
  .stat-tip { font-size: 12px; color: #909399; margin-left: 4px; }
</style>
