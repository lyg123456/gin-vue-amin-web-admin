<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex gap-2 flex-wrap items-center">
        <el-button type="primary" icon="plus" @click="openDrawer">新增</el-button>
        <el-input
          v-model="keyword"
          style="width: 240px"
          placeholder="标题/Slug 搜索"
          clearable
          @clear="getTableData"
          @keyup.enter="getTableData"
        />
        <el-select v-model="status" style="width: 160px" clearable placeholder="状态" @change="getTableData">
          <el-option label="草稿" value="draft" />
          <el-option label="已发布" value="published" />
          <el-option label="已归档" value="archived" />
        </el-select>
      </div>

      <el-table :data="tableData" style="width: 100%" row-key="ID">
        <el-table-column align="left" label="创建时间" width="180">
          <template #default="scope">
            <span>{{ formatDate(scope.row.CreatedAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="标题" prop="title" min-width="220" />
        <el-table-column align="left" label="Slug" prop="slug" min-width="200" />
        <el-table-column align="left" label="状态" width="110">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'published'" type="success">已发布</el-tag>
            <el-tag v-else-if="scope.row.status === 'archived'" type="info">已归档</el-tag>
            <el-tag v-else type="warning">草稿</el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="浏览量" prop="viewCount" width="90" />
        <el-table-column align="left" label="操作" min-width="220">
          <template #default="scope">
            <el-button type="primary" link icon="edit" @click="editArticle(scope.row)">编辑</el-button>
            <el-button
              v-if="scope.row.status !== 'published'"
              type="primary"
              link
              icon="promotion"
              @click="publishArticle(scope.row)"
            >
              发布
            </el-button>
            <el-button type="primary" link icon="delete" @click="removeArticle(scope.row)">删除</el-button>
            <el-button type="primary" link icon="link" @click="copyPublicUrl(scope.row)">复制链接</el-button>
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

    <el-drawer v-model="drawerVisible" :before-close="closeDrawer" :show-close="false" size="60%">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ drawerType === 'update' ? '编辑文章' : '新增文章' }}</span>
          <div>
            <el-button @click="closeDrawer">取消</el-button>
            <el-button type="primary" @click="submitDrawer">保存</el-button>
          </div>
        </div>
      </template>

      <el-form :model="form" label-width="110px">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="建议 20-40 字" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" placeholder="例如：how-to-get-customers" />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="form.summary" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="配图">
          <SelectImage v-model="galleryUrls" multiple :max-update-count="6" file-type="image" />
          <p class="field-tip">最多 6 张，对应库字段 <code>cover_image</code>（多图逗号分隔保存）</p>
        </el-form-item>
        <el-form-item label="SEO 标题">
          <el-input v-model="form.seoTitle" placeholder="留空则可用标题替代（后续可优化）" />
        </el-form-item>
        <el-form-item label="SEO 关键词">
          <el-input v-model="form.seoKeywords" placeholder="用逗号分隔，例如：获客,SEO,中小企业" />
        </el-form-item>
        <el-form-item label="SEO 描述">
          <el-input v-model="form.seoDescription" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="正文(Markdown)">
          <el-input v-model="form.content" type="textarea" :rows="14" placeholder="先用 Markdown 文本，后续可接富文本编辑器" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import {
    createContentArticle,
    updateContentArticle,
    deleteContentArticle,
    findContentArticle,
    getContentArticleList,
    publishContentArticle
  } from '@/api/contentArticle'

  defineOptions({
    name: 'ContentArticle'
  })

  const keyword = ref('')
  const status = ref('')

  const page = ref(1)
  const total = ref(0)
  const pageSize = ref(10)
  const tableData = ref([])

  const handleSizeChange = (val) => {
    pageSize.value = val
    getTableData()
  }

  const handleCurrentChange = (val) => {
    page.value = val
    getTableData()
  }

  const getTableData = async () => {
    const res = await getContentArticleList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value,
      status: status.value
    })
    if (res.code === 0) {
      tableData.value = res.data.list
      total.value = res.data.total
      page.value = res.data.page
      pageSize.value = res.data.pageSize
    }
  }

  getTableData()

  const drawerVisible = ref(false)
  const drawerType = ref('create')

  const emptyForm = () => ({
    title: '',
    slug: '',
    summary: '',
    content: '',
    contentType: 'markdown',
    coverImage: '',
    seoTitle: '',
    seoKeywords: '',
    seoDescription: '',
    status: 'draft'
  })

  const form = ref(emptyForm())
  /** 多图上传，提交时写入 form.coverImage（逗号分隔） */
  const galleryUrls = ref([])

  const parseCoverToGallery = (cover) => {
    if (!cover) return []
    return String(cover)
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
      .slice(0, 6)
  }

  const openDrawer = () => {
    drawerType.value = 'create'
    form.value = emptyForm()
    galleryUrls.value = []
    drawerVisible.value = true
  }

  const closeDrawer = () => {
    drawerVisible.value = false
    form.value = emptyForm()
    galleryUrls.value = []
  }

  const editArticle = async (row) => {
    const res = await findContentArticle({ id: row.ID })
    if (res.code === 0) {
      drawerType.value = 'update'
      form.value = res.data
      galleryUrls.value = parseCoverToGallery(res.data.coverImage)
      drawerVisible.value = true
    }
  }

  const submitDrawer = async () => {
    const urls = Array.isArray(galleryUrls.value) ? galleryUrls.value.slice(0, 6) : []
    const payload = { ...form.value, coverImage: urls.join(',') }
    let res
    if (drawerType.value === 'update') {
      res = await updateContentArticle(payload)
    } else {
      res = await createContentArticle(payload)
    }
    if (res.code === 0) {
      ElMessage.success('保存成功')
      closeDrawer()
      getTableData()
    }
  }

  const removeArticle = async (row) => {
    ElMessageBox.confirm('确定要删除这篇文章吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await deleteContentArticle({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        if (tableData.value.length === 1 && page.value > 1) page.value--
        getTableData()
      }
    })
  }

  const publishArticle = async (row) => {
    ElMessageBox.confirm('发布后即可通过公开链接访问（用于 SEO 收录）。确定发布吗?', '提示', {
      confirmButtonText: '发布',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
      const res = await publishContentArticle({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('发布成功')
        getTableData()
      }
    })
  }

  const copyPublicUrl = async (row) => {
    const url = `${window.location.origin}/web/#/article/${row.slug}`
    try {
      await navigator.clipboard.writeText(url)
      ElMessage.success('已复制公开链接')
    } catch (e) {
      ElMessage.warning('复制失败，请手动复制')
    }
  }
</script>

<style scoped>
  .field-tip {
    margin: 8px 0 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.5;
  }
  .field-tip code {
    font-size: 12px;
    background: #f4f4f5;
    padding: 1px 6px;
    border-radius: 4px;
  }
</style>

