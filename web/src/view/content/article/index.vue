<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex gap-2 flex-wrap items-center">
        <el-button type="primary" icon="plus" @click="openDrawer">新增</el-button>
        <el-button type="default" icon="folder-opened" @click="openCategoryDialog">分类管理</el-button>
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
        <el-table-column align="left" label="分类" min-width="120">
          <template #default="scope">
            <span>{{ scope.row.category?.name || '—' }}</span>
          </template>
        </el-table-column>
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
        <el-form-item label="分类">
          <el-cascader
            v-model="categoryPick"
            :options="categoryTreeOpts"
            :props="cascaderProps"
            clearable
            filterable
            placeholder="选择一级或二级分类（最多二级）"
            style="width: 100%"
            @change="onCategoryCascaderChange"
          />
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

    <el-dialog v-model="categoryDialogVisible" title="文章分类" width="720px" destroy-on-close @closed="onCategoryDialogClosed">
      <div class="mb-3">
        <el-button type="primary" size="small" icon="plus" @click="openCategoryForm('create')">新增分类</el-button>
      </div>
      <el-table :data="categoryFlatList" border size="small" max-height="420">
        <el-table-column prop="ID" label="ID" width="70" />
        <el-table-column label="父级" min-width="100">
          <template #default="{ row }">
            <span>{{ parentName(row.parentId) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openCategoryForm('edit', row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="removeCategory(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="categoryFormVisible" :title="categoryFormType === 'edit' ? '编辑分类' : '新增分类'" width="480px" destroy-on-close>
      <el-form :model="categoryForm" label-width="90px">
        <el-form-item label="父级">
          <el-select v-model="categoryForm.parentId" :disabled="categoryFormType === 'edit'" placeholder="顶级为一级分类" style="width: 100%">
            <el-option label="（顶级）" :value="0" />
            <el-option v-for="r in rootCategories" :key="r.ID" :label="r.name" :value="r.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="categoryForm.name" maxlength="100" show-word-limit placeholder="分类名称" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="categoryForm.sort" :min="0" :max="9999" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="categoryForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryFormVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCategoryForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
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
  import {
    getContentArticleCategoryTree,
    getContentArticleCategoryList,
    createContentArticleCategory,
    updateContentArticleCategory,
    deleteContentArticleCategory
  } from '@/api/contentArticleCategory'

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

  const cascaderProps = {
    value: 'value',
    label: 'label',
    children: 'children',
    checkStrictly: true,
    emitPath: false
  }
  const categoryTreeOpts = ref([])
  const categoryPick = ref()
  const categoryFlatList = ref([])
  const categoryDialogVisible = ref(false)
  const categoryFormVisible = ref(false)
  const categoryFormType = ref('create')
  const categoryForm = ref({
    ID: 0,
    parentId: 0,
    name: '',
    sort: 0,
    enabled: true
  })

  const rootCategories = computed(() => (categoryFlatList.value || []).filter((c) => c.parentId === 0))

  const parentName = (pid) => {
    if (!pid) return '—'
    const p = categoryFlatList.value.find((x) => x.ID === pid)
    return p?.name || pid
  }

  const loadCategoryTree = async () => {
    const res = await getContentArticleCategoryTree()
    if (res.code === 0) {
      categoryTreeOpts.value = res.data?.list || []
    }
  }

  const loadCategoryFlat = async () => {
    const res = await getContentArticleCategoryList()
    if (res.code === 0) {
      categoryFlatList.value = res.data?.list || []
    }
  }

  const syncCategoryPickFromForm = () => {
    const id = Number(form.value.categoryId || 0)
    categoryPick.value = id > 0 ? id : undefined
  }

  const onCategoryCascaderChange = (v) => {
    form.value.categoryId = v ? Number(v) : 0
  }

  const openCategoryDialog = async () => {
    await loadCategoryFlat()
    categoryDialogVisible.value = true
  }

  const onCategoryDialogClosed = () => {
    loadCategoryTree()
  }

  const openCategoryForm = (type, row) => {
    categoryFormType.value = type
    if (type === 'create') {
      categoryForm.value = { ID: 0, parentId: 0, name: '', sort: 0, enabled: true }
    } else if (row) {
      categoryForm.value = {
        ID: row.ID,
        parentId: row.parentId,
        name: row.name,
        sort: row.sort,
        enabled: !!row.enabled
      }
    }
    categoryFormVisible.value = true
  }

  const submitCategoryForm = async () => {
    const f = categoryForm.value
    if (!f.name?.trim()) {
      ElMessage.warning('请填写分类名称')
      return
    }
    let res
    if (categoryFormType.value === 'edit') {
      res = await updateContentArticleCategory({
        ID: f.ID,
        parentId: f.parentId,
        name: f.name.trim(),
        sort: f.sort,
        enabled: f.enabled
      })
    } else {
      res = await createContentArticleCategory({
        parentId: f.parentId,
        name: f.name.trim(),
        sort: f.sort,
        enabled: f.enabled
      })
    }
    if (res.code === 0) {
      ElMessage.success('保存成功')
      categoryFormVisible.value = false
      await loadCategoryFlat()
      await loadCategoryTree()
    }
  }

  const removeCategory = (row) => {
    ElMessageBox.confirm(`确定删除分类「${row.name}」？`, '提示', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    }).then(async () => {
      const res = await deleteContentArticleCategory({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        await loadCategoryFlat()
        await loadCategoryTree()
      }
    })
  }

  loadCategoryTree()

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
    status: 'draft',
    categoryId: 0
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
    syncCategoryPickFromForm()
    drawerVisible.value = true
  }

  const closeDrawer = () => {
    drawerVisible.value = false
    form.value = emptyForm()
    galleryUrls.value = []
    categoryPick.value = undefined
  }

  const editArticle = async (row) => {
    const res = await findContentArticle({ id: row.ID })
    if (res.code === 0) {
      drawerType.value = 'update'
      form.value = res.data
      galleryUrls.value = parseCoverToGallery(res.data.coverImage)
      syncCategoryPickFromForm()
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

