<template>
  <div class="gva-table-box">
    <div class="gva-btn-list flex gap-2 flex-wrap items-center">
      <el-button type="primary" icon="plus" @click="$router.push({ name: 'shortVideoAi' })">AI 生成短视频</el-button>
      <el-input v-model="keyword" style="width: 220px" clearable placeholder="标题搜索" @keyup.enter="load" @clear="load" />
      <el-select v-model="status" style="width: 140px" clearable placeholder="状态" @change="load">
        <el-option label="草稿" value="draft" />
        <el-option label="生成中" value="generating" />
        <el-option label="就绪" value="ready" />
        <el-option label="失败" value="failed" />
        <el-option label="已发布" value="published" />
      </el-select>
      <el-button @click="load">刷新</el-button>
    </div>

    <el-table :data="tableData" row-key="ID" style="width: 100%">
      <el-table-column label="ID" prop="ID" width="70" />
      <el-table-column label="标题" prop="title" min-width="160" show-overflow-tooltip />
      <el-table-column label="时长" width="80">
        <template #default="{ row }">{{ row.durationSec }}s</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="播放" prop="viewCount" width="80" />
      <el-table-column label="成片" width="88">
        <template #default="{ row }">
          <el-tag v-if="row.videoUrl" type="success" size="small">有</el-tag>
          <el-tag v-else type="info" size="small">无</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="openDetail(row)">详情</el-button>
          <el-button v-if="row.status !== 'published'" type="primary" link @click="onPublish(row)">发布</el-button>
          <el-button type="danger" link @click="onDelete(row)">删除</el-button>
        </template>
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

    <el-drawer v-model="drawerVisible" size="56%" :title="drawerTitle">
      <el-form v-if="detail" :model="detail" label-width="110px">
        <el-form-item label="标题"><el-input v-model="detail.title" /></el-form-item>
        <el-form-item label="Slug"><el-input v-model="detail.slug" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="detail.description" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="创意文字"><el-input v-model="detail.promptText" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="时长(秒)">
          <el-input-number v-model="detail.durationSec" :min="30" :max="120" />
        </el-form-item>
        <el-form-item label="脚本">
          <el-input v-model="detail.script" type="textarea" :rows="10" />
        </el-form-item>
        <el-form-item label="封面">
          <SelectImage v-model="detailCover" file-type="image" />
        </el-form-item>
        <el-form-item label="首帧图">
          <SelectImage v-model="detailFirstFrame" file-type="image" />
        </el-form-item>
        <el-form-item label="尾帧图">
          <SelectImage v-model="detailLastFrame" file-type="image" />
        </el-form-item>
        <el-form-item label="成片">
          <SelectImage v-model="detailVideo" file-type="video" />
        </el-form-item>
        <el-form-item v-if="detail.generationError" label="失败原因">
          <el-text type="danger">{{ detail.generationError }}</el-text>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="saveDetail">保存</el-button>
          <el-button :loading="regenScriptLoading" @click="regen(true, false)">重生成脚本</el-button>
          <el-button :loading="regenVideoLoading" @click="regen(false, true)">重生成成片</el-button>
        </el-form-item>
      </el-form>
      <div v-if="backendPlayUrl" class="player-wrap">
        <p class="play-label">成片预览</p>
        <video :key="backendPlayUrl" :src="backendPlayUrl" controls class="video-player" />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import { videoPlayUrl } from '@/utils/image'
  import SelectImage from '@/components/selectImage/selectImage.vue'
  import {
    getShortVideoList,
    findShortVideo,
    updateShortVideo,
    deleteShortVideo,
    publishShortVideo,
    regenerateShortVideo
  } from '@/api/contentShortVideo'

  defineOptions({ name: 'ShortVideoList' })

  const keyword = ref('')
  const status = ref('')
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const tableData = ref([])

  const drawerVisible = ref(false)
  const detail = ref(null)
  const detailCover = ref('')
  const detailVideo = ref('')
  const detailFirstFrame = ref('')
  const detailLastFrame = ref('')

  const pickImageUrl = (val) => {
    if (Array.isArray(val)) return (val[0] || '').trim()
    return (val || '').trim()
  }
  const saving = ref(false)
  const regenScriptLoading = ref(false)
  const regenVideoLoading = ref(false)

  const drawerTitle = computed(() => (detail.value?.title ? `短视频：${detail.value.title}` : '短视频详情'))
  /** 播放器只读后端 videoUrl，不与 SelectImage 编辑态混用 */
  const backendPlayUrl = computed(() => videoPlayUrl(detail.value?.videoUrl))

  const syncDetailFromBackend = (data) => {
    if (!data) return
    detail.value = { ...data }
    detailCover.value = data.coverImage || ''
    detailVideo.value = data.videoUrl || ''
    const urls = (data.sourceImages || '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    detailFirstFrame.value = data.firstFrameUrl || urls[0] || ''
    detailLastFrame.value = data.lastFrameUrl || urls[1] || ''
  }

  const statusLabel = (s) =>
    ({
      draft: '草稿',
      generating: '生成中',
      ready: '就绪',
      failed: '失败',
      published: '已发布',
      archived: '已归档'
    })[s] || s

  const statusType = (s) =>
    ({ published: 'success', ready: 'success', generating: 'warning', failed: 'danger' })[s] || 'info'

  const load = async () => {
    const res = await getShortVideoList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value,
      status: status.value
    })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total
    }
  }

  const onPage = (p) => {
    page.value = p
    load()
  }

  load()

  const openDetail = async (row) => {
    const res = await findShortVideo({ id: row.ID })
    if (res.code === 0) {
      syncDetailFromBackend(res.data)
      drawerVisible.value = true
    }
  }

  const saveDetail = async () => {
    if (!detail.value) return
    saving.value = true
    try {
      const cover = pickImageUrl(detailCover.value)
      const video = pickImageUrl(detailVideo.value)
      const first = pickImageUrl(detailFirstFrame.value)
      const last = pickImageUrl(detailLastFrame.value)
      const payload = {
        ...detail.value,
        coverImage: cover,
        videoUrl: video,
        firstFrameUrl: first,
        lastFrameUrl: last
      }
      const res = await updateShortVideo(payload)
      if (res.code === 0) {
        detail.value = { ...detail.value, ...payload }
        detailVideo.value = video
        ElMessage.success('已保存')
        load()
        const fresh = await findShortVideo({ id: detail.value.ID })
        if (fresh.code === 0) syncDetailFromBackend(fresh.data)
      }
    } finally {
      saving.value = false
    }
  }

  const regen = async (regenScript, regenVideo) => {
    if (!detail.value?.ID) return
    if (regenVideo) {
      const first = pickImageUrl(detailFirstFrame.value)
      const last = pickImageUrl(detailLastFrame.value)
      if (!first || !last) {
        ElMessage.warning('重生成成片需同时配置首帧图与尾帧图')
        return
      }
    }
    if (regenScript) regenScriptLoading.value = true
    if (regenVideo) regenVideoLoading.value = true
    try {
      const res = await regenerateShortVideo({
        id: detail.value.ID,
        title: detail.value.title,
        description: detail.value.description,
        promptText: detail.value.promptText,
        script: detail.value.script,
        durationSec: detail.value.durationSec,
        coverImage: pickImageUrl(detailCover.value),
        firstFrameUrl: pickImageUrl(detailFirstFrame.value),
        lastFrameUrl: pickImageUrl(detailLastFrame.value),
        regenScript,
        regenVideo
      })
      if (res.code === 0) {
        syncDetailFromBackend(res.data)
        ElMessage.success('操作成功')
        load()
      }
    } finally {
      regenScriptLoading.value = false
      regenVideoLoading.value = false
    }
  }

  const onPublish = (row) => {
    ElMessageBox.confirm('发布后将在用户端短视频列表展示，确定发布？', '提示', { type: 'warning' }).then(async () => {
      const res = await publishShortVideo({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已发布')
        load()
      }
    })
  }

  const onDelete = (row) => {
    ElMessageBox.confirm('确定删除？', '提示', { type: 'warning' }).then(async () => {
      const res = await deleteShortVideo({ id: row.ID })
      if (res.code === 0) {
        ElMessage.success('已删除')
        load()
      }
    })
  }
</script>

<style scoped>
  .player-wrap { margin-top: 16px; }
  .play-label { font-size: 12px; color: #909399; margin: 0 0 8px; }
  .video-player { width: 100%; max-height: 360px; background: #000; border-radius: 8px; }
</style>
