<template>
  <div class="tool-panel speed-tool">
    <el-button type="primary" :loading="running" @click="runFullTest">开始完整测速</el-button>
    <el-button :disabled="running" @click="runLatencyOnly">仅测延迟</el-button>

    <el-row :gutter="12" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="stat-label">延迟 (平均)</div>
          <div class="stat-value">{{ fmtMs(latency.avg) }}</div>
          <div class="stat-sub">抖动 {{ fmtMs(latency.jitter) }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="stat-label">下载速度</div>
          <div class="stat-value">{{ fmtMbps(download.mbps) }}</div>
          <div class="stat-sub">{{ download.mb }}MB · {{ fmtMs(download.ms) }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="stat-label">上传速度</div>
          <div class="stat-value">{{ fmtMbps(upload.mbps) }}</div>
          <div class="stat-sub">{{ upload.mb }}MB · {{ fmtMs(upload.ms) }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <div class="stat-card">
          <div class="stat-label">网络类型</div>
          <div class="stat-value stat-value-sm">{{ conn.effectiveType || '—' }}</div>
          <div class="stat-sub">{{ conn.downlink ? `下行 ${conn.downlink} Mbps` : '浏览器未提供' }}</div>
        </div>
      </el-col>
    </el-row>

    <el-progress v-if="running" :percentage="progress" :stroke-width="10" class="mt-3" />

    <el-descriptions title="延迟明细" :column="2" border class="mt-3" size="small">
      <el-descriptions-item label="最小">{{ fmtMs(latency.min) }}</el-descriptions-item>
      <el-descriptions-item label="最大">{{ fmtMs(latency.max) }}</el-descriptions-item>
      <el-descriptions-item label="平均">{{ fmtMs(latency.avg) }}</el-descriptions-item>
      <el-descriptions-item label="抖动">{{ fmtMs(latency.jitter) }}</el-descriptions-item>
      <el-descriptions-item label="采样次数">{{ latency.samples.length }} 次</el-descriptions-item>
      <el-descriptions-item label="服务器节点">{{ serverInfo.node || '—' }}</el-descriptions-item>
    </el-descriptions>

    <el-descriptions title="环境与出口" :column="1" border class="mt-3" size="small">
      <el-descriptions-item label="出口 IP">{{ serverInfo.clientIP || '—' }}</el-descriptions-item>
      <el-descriptions-item label="User-Agent">{{ uaShort }}</el-descriptions-item>
      <el-descriptions-item label="RTT / 下行(浏览器)">
        {{ conn.rtt != null ? `${conn.rtt} ms` : '—' }} /
        {{ conn.downlink != null ? `${conn.downlink} Mbps` : '—' }}
      </el-descriptions-item>
      <el-descriptions-item label="说明">
        下载/上传测速走本站 API；延迟为浏览器到服务器的 HTTP 往返。结果受本地 WiFi、代理、服务器带宽影响。
      </el-descriptions-item>
    </el-descriptions>

    <el-table v-if="latency.samples.length" :data="latencyTable" size="small" class="mt-3" max-height="200">
      <el-table-column prop="i" label="#" width="50" />
      <el-table-column prop="ms" label="延迟 (ms)" />
    </el-table>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { speedInfo, speedPing } from '@/api/publicPortalOffice'

  const apiBase = import.meta.env.VITE_BASE_API || ''

  const running = ref(false)
  const progress = ref(0)
  const serverInfo = reactive({ clientIP: '', node: '' })
  const latency = reactive({ min: 0, max: 0, avg: 0, jitter: 0, samples: [] })
  const download = reactive({ mbps: 0, mb: 0, ms: 0 })
  const upload = reactive({ mbps: 0, mb: 0, ms: 0 })
  const conn = reactive({ effectiveType: '', downlink: null, rtt: null })

  const uaShort = computed(() => {
    const ua = navigator.userAgent || ''
    return ua.length > 120 ? ua.slice(0, 120) + '…' : ua
  })

  const latencyTable = computed(() =>
    latency.samples.map((ms, i) => ({ i: i + 1, ms: ms.toFixed(1) }))
  )

  const fmtMs = (v) => (v > 0 ? `${v.toFixed(1)} ms` : '—')
  const fmtMbps = (v) => (v > 0 ? `${v.toFixed(2)} Mbps` : '—')

  const readConnection = () => {
    const n = navigator.connection || navigator.mozConnection || navigator.webkitConnection
    if (!n) return
    conn.effectiveType = n.effectiveType || ''
    conn.downlink = n.downlink ?? null
    conn.rtt = n.rtt ?? null
  }

  const pingOnce = async () => {
    const t0 = performance.now()
    await speedPing()
    return performance.now() - t0
  }

  const runLatency = async (n = 10) => {
    const samples = []
    for (let i = 0; i < n; i++) {
      samples.push(await pingOnce())
      progress.value = Math.min(30, 3 * (i + 1))
    }
    latency.samples = samples
    latency.min = Math.min(...samples)
    latency.max = Math.max(...samples)
    latency.avg = samples.reduce((a, b) => a + b, 0) / samples.length
    const diffs = samples.map((s) => Math.abs(s - latency.avg))
    latency.jitter = diffs.reduce((a, b) => a + b, 0) / diffs.length
  }

  const runDownload = async (mb) => {
    const url = `${apiBase}/public/office/speed/download?mb=${mb}&_t=${Date.now()}`
    const t0 = performance.now()
    const res = await fetch(url, { cache: 'no-store' })
    const blob = await res.blob()
    const ms = performance.now() - t0
    const bits = blob.size * 8
    download.mb = mb
    download.ms = ms
    download.mbps = ms > 0 ? bits / ms / 1000 : 0
  }

  const runUpload = async (mb) => {
    const bytes = mb * 1024 * 1024
    const buf = new Uint8Array(bytes)
    const t0 = performance.now()
    const res = await fetch(`${apiBase}/public/office/speed/upload`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/octet-stream' },
      body: buf
    })
    const json = await res.json()
    const ms = performance.now() - t0
    if (json.code !== 0) throw new Error(json.msg || '上传测速失败')
    const got = json.data?.bytes || bytes
    upload.mb = mb
    upload.ms = ms
    upload.mbps = ms > 0 ? (got * 8) / ms / 1000 : 0
  }

  const runFullTest = async () => {
    running.value = true
    progress.value = 0
    try {
      readConnection()
      const info = await speedInfo()
      const infoData = info.data?.data ?? info.data
      if (infoData) {
        serverInfo.clientIP = infoData.clientIP
        serverInfo.node = infoData.node
      }
      await runLatency(10)
      progress.value = 40
      await runDownload(1)
      progress.value = 65
      await runDownload(5)
      progress.value = 80
      await runUpload(2)
      progress.value = 100
      ElMessage.success('测速完成')
    } catch (e) {
      ElMessage.error(e?.message || '测速失败')
    } finally {
      running.value = false
    }
  }

  const runLatencyOnly = async () => {
    running.value = true
    try {
      await runLatency(12)
      ElMessage.success('延迟测试完成')
    } catch (e) {
      ElMessage.error(e?.message || '失败')
    } finally {
      running.value = false
      progress.value = 0
    }
  }

  onMounted(() => {
    readConnection()
    speedInfo().then((res) => {
      const d = res.data?.data ?? res.data
      if (d) {
        serverInfo.clientIP = d.clientIP
        serverInfo.node = d.node
      }
    })
  })
</script>

<style scoped>
  .speed-tool .stat-row {
    margin-top: 16px;
  }
  .stat-card {
    background: #f8fafc;
    border: 1px solid #e8e8e8;
    border-radius: 4px;
    padding: 12px 14px;
    margin-bottom: 10px;
    min-height: 88px;
  }
  .stat-label {
    font-size: 13px;
    color: #888;
  }
  .stat-value {
    font-size: 22px;
    font-weight: 700;
    color: var(--portal-brand, #1a73e8);
    margin: 6px 0 4px;
  }
  .stat-value-sm {
    font-size: 16px;
  }
  .stat-sub {
    font-size: 12px;
    color: #666;
  }
  .mt-3 {
    margin-top: 14px;
  }
</style>
