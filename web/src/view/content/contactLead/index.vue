<template>
  <div class="gva-table-box">
    <div class="gva-btn-list">
      <el-input
        v-model="search.keyword"
        clearable
        placeholder="电话 / 备注关键字"
        style="width: 220px"
        @keyup.enter="getTableData"
      />
      <el-button type="primary" icon="search" @click="getTableData">查询</el-button>
    </div>
    <el-table :data="tableData" row-key="ID">
      <el-table-column align="left" label="ID" min-width="60" prop="ID" />
      <el-table-column align="left" label="电话" min-width="120" prop="phone" />
      <el-table-column align="left" label="备注" min-width="200" prop="remark" show-overflow-tooltip />
      <el-table-column align="left" label="IP" min-width="120" prop="clientIp" />
      <el-table-column align="left" label="提交时间" min-width="160" prop="CreatedAt">
        <template #default="{ row }">
          {{ formatTime(row.CreatedAt) }}
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
  import { getPortalContactLeadList } from '@/api/contentPortalContactLead'
  import { ref, onMounted } from 'vue'

  defineOptions({
    name: 'ContentPortalContactLead'
  })

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])
  const search = ref({ keyword: '' })

  const formatTime = (t) => {
    if (!t) return '--'
    const d = new Date(t)
    if (Number.isNaN(d.getTime())) return '--'
    return d.toLocaleString('zh-CN')
  }

  const getTableData = async () => {
    const res = await getPortalContactLeadList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: search.value.keyword || undefined
    })
    if (res.code === 0 && res.data) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      page.value = res.data.page || 1
      pageSize.value = res.data.pageSize || 10
    }
  }

  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  onMounted(() => {
    getTableData()
  })
</script>

<style scoped lang="scss"></style>
