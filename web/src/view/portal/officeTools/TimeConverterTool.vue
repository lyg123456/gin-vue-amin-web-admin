<template>
  <div class="tool-panel">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="时间戳" name="ts">
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-input v-model="timestampInput" placeholder="秒级或毫秒级时间戳" clearable @input="onTimestampInput">
              <template #append>
                <el-button @click="useNow">当前</el-button>
              </template>
            </el-input>
            <p class="hint">13 位为毫秒，10 位为秒</p>
            <el-input v-model="datetimeLocal" type="datetime-local" class="mt-3" @change="onDatetimeChange" />
          </el-col>
          <el-col :xs="24" :md="12">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="本地">{{ result.local }}</el-descriptions-item>
              <el-descriptions-item label="UTC">{{ result.utc }}</el-descriptions-item>
              <el-descriptions-item label="ISO 8601">{{ result.iso }}</el-descriptions-item>
              <el-descriptions-item label="秒">{{ result.sec }}</el-descriptions-item>
              <el-descriptions-item label="毫秒">{{ result.ms }}</el-descriptions-item>
              <el-descriptions-item label="星期">{{ result.weekday }}</el-descriptions-item>
              <el-descriptions-item label="相对">{{ result.relative }}</el-descriptions-item>
            </el-descriptions>
            <el-button size="small" class="mt-2" @click="copyText(result.iso)">复制 ISO</el-button>
          </el-col>
        </el-row>
      </el-tab-pane>

      <el-tab-pane label="时区转换" name="tz">
        <el-row :gutter="12" class="mb-2">
          <el-col :span="24">
            <el-input v-model="tzInput" type="datetime-local" @change="calcTimezone" />
          </el-col>
          <el-col :xs="24" :md="12">
            <el-select v-model="tzFrom" filterable class="w-full" @change="calcTimezone">
              <el-option v-for="z in timezones" :key="'f-' + z" :label="z" :value="z" />
            </el-select>
            <p class="hint">源时区</p>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-select v-model="tzTo" filterable class="w-full" @change="calcTimezone">
              <el-option v-for="z in timezones" :key="'t-' + z" :label="z" :value="z" />
            </el-select>
            <p class="hint">目标时区</p>
          </el-col>
        </el-row>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="tzFrom">{{ tzResult.from }}</el-descriptions-item>
          <el-descriptions-item :label="tzTo">{{ tzResult.to }}</el-descriptions-item>
          <el-descriptions-item label="UTC">{{ tzResult.utc }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <el-tab-pane label="时间差" name="diff">
        <el-row :gutter="12">
          <el-col :xs="24" :md="12">
            <span class="label">开始</span>
            <el-input v-model="diffStart" type="datetime-local" @change="calcDiff" />
          </el-col>
          <el-col :xs="24" :md="12">
            <span class="label">结束</span>
            <el-input v-model="diffEnd" type="datetime-local" @change="calcDiff" />
          </el-col>
        </el-row>
        <el-button class="mt-2" size="small" @click="swapDiff">交换起止</el-button>
        <el-descriptions :column="1" border size="small" class="mt-3">
          <el-descriptions-item label="相差">{{ diffResult.total }}</el-descriptions-item>
          <el-descriptions-item label="天">{{ diffResult.days }}</el-descriptions-item>
          <el-descriptions-item label="时:分:秒">{{ diffResult.hms }}</el-descriptions-item>
          <el-descriptions-item label="工作日(近似)">{{ diffResult.workdays }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <el-tab-pane label="批量时间戳" name="batch">
        <el-input v-model="batchInput" type="textarea" :rows="6" placeholder="每行一个时间戳（秒或毫秒）" />
        <el-button type="primary" class="mt-2" @click="runBatch">转换</el-button>
        <el-table :data="batchRows" size="small" class="mt-3" empty-text="暂无结果">
          <el-table-column prop="raw" label="输入" width="140" />
          <el-table-column prop="local" label="本地时间" min-width="180" />
          <el-table-column prop="iso" label="ISO" min-width="200" />
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="Cron 说明" name="cron">
        <el-alert
          title="常见 Cron（Linux）：分 时 日 月 周。例：0 0 * * * 每天零点；*/5 * * * * 每 5 分钟。"
          type="info"
          :closable="false"
          show-icon
        />
        <el-input v-model="cronExpr" class="mt-3" placeholder="输入 cron 表达式，如 0 9 * * 1-5" />
        <p class="hint mt-2">{{ cronHint }}</p>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  import { computed, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'

  const activeTab = ref('ts')
  const timestampInput = ref('')
  const datetimeLocal = ref('')
  const result = reactive({
    local: '--',
    utc: '--',
    iso: '--',
    sec: '--',
    ms: '--',
    weekday: '--',
    relative: '--'
  })

  const pad = (n) => String(n).padStart(2, '0')
  const formatLocal = (d) =>
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  const toDatetimeLocalValue = (d) =>
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`

  const relativeText = (d) => {
    const sec = Math.round((d.getTime() - Date.now()) / 1000)
    const abs = Math.abs(sec)
    const unit =
      abs < 60 ? `${abs} 秒` : abs < 3600 ? `${Math.floor(abs / 60)} 分钟` : abs < 86400 ? `${Math.floor(abs / 3600)} 小时` : `${Math.floor(abs / 86400)} 天`
    return sec >= 0 ? `${unit}后` : `${unit}前`
  }

  const weekdays = ['日', '一', '二', '三', '四', '五', '六']

  const fillFromDate = (d) => {
    if (Number.isNaN(d.getTime())) return
    result.local = formatLocal(d)
    result.utc = d.toISOString().replace('T', ' ').replace('Z', ' UTC')
    result.iso = d.toISOString()
    result.sec = String(Math.floor(d.getTime() / 1000))
    result.ms = String(d.getTime())
    result.weekday = `星期${weekdays[d.getDay()]}`
    result.relative = relativeText(d)
    datetimeLocal.value = toDatetimeLocalValue(d)
    timestampInput.value = result.sec
  }

  const onTimestampInput = () => {
    const raw = String(timestampInput.value || '').trim()
    if (!raw || !/^\d+$/.test(raw)) return
    let ms = Number(raw)
    if (raw.length <= 10) ms *= 1000
    fillFromDate(new Date(ms))
  }

  const onDatetimeChange = () => {
    if (!datetimeLocal.value) return
    fillFromDate(new Date(datetimeLocal.value))
  }

  const useNow = () => fillFromDate(new Date())

  const copyText = async (t) => {
    await navigator.clipboard.writeText(t)
    ElMessage.success('已复制')
  }

  // 时区
  const timezones = Intl.supportedValuesOf('timeZone')
  const tzInput = ref(toDatetimeLocalValue(new Date()))
  const tzFrom = ref(Intl.DateTimeFormat().resolvedOptions().timeZone)
  const tzTo = ref('UTC')
  const tzResult = reactive({ from: '--', to: '--', utc: '--' })

  const formatInZone = (d, zone) => {
    return new Intl.DateTimeFormat('zh-CN', {
      timeZone: zone,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    }).format(d)
  }

  const calcTimezone = () => {
    if (!tzInput.value) return
    const d = new Date(tzInput.value)
    if (Number.isNaN(d.getTime())) return
    tzResult.from = formatInZone(d, tzFrom.value)
    tzResult.to = formatInZone(d, tzTo.value)
    tzResult.utc = d.toISOString()
  }

  // 时间差
  const diffStart = ref(toDatetimeLocalValue(new Date(Date.now() - 86400000)))
  const diffEnd = ref(toDatetimeLocalValue(new Date()))
  const diffResult = reactive({ total: '--', days: '--', hms: '--', workdays: '--' })

  const calcDiff = () => {
    const a = new Date(diffStart.value).getTime()
    const b = new Date(diffEnd.value).getTime()
    if (Number.isNaN(a) || Number.isNaN(b)) return
    let ms = b - a
    const sign = ms < 0 ? '-' : ''
    ms = Math.abs(ms)
    const days = Math.floor(ms / 86400000)
    const rem = ms % 86400000
    const h = Math.floor(rem / 3600000)
    const m = Math.floor((rem % 3600000) / 60000)
    const s = Math.floor((rem % 60000) / 1000)
    diffResult.days = String(days)
    diffResult.hms = `${pad(h)}:${pad(m)}:${pad(s)}`
    diffResult.total = `${sign}${days} 天 ${diffResult.hms}`
    diffResult.workdays = String(Math.floor(days * 5 / 7))
  }

  const swapDiff = () => {
    const t = diffStart.value
    diffStart.value = diffEnd.value
    diffEnd.value = t
    calcDiff()
  }

  // 批量
  const batchInput = ref('')
  const batchRows = ref([])
  const runBatch = () => {
    const lines = batchInput.value.split(/\n/).map((l) => l.trim()).filter(Boolean)
    batchRows.value = lines.map((raw) => {
      let ms = Number(raw)
      if (raw.length <= 10) ms *= 1000
      const d = new Date(ms)
      return {
        raw,
        local: Number.isNaN(d.getTime()) ? '无效' : formatLocal(d),
        iso: Number.isNaN(d.getTime()) ? '-' : d.toISOString()
      }
    })
  }

  const cronExpr = ref('0 9 * * 1-5')
  const cronHint = computed(() => {
    const e = cronExpr.value.trim()
    if (!e) return ''
    const parts = e.split(/\s+/)
    if (parts.length < 5) return '表达式字段不足，标准 5 段：分 时 日 月 周'
    return `已解析 ${parts.length} 段：分=${parts[0]} 时=${parts[1]} 日=${parts[2]} 月=${parts[3]} 周=${parts[4]}`
  })

  useNow()
  calcTimezone()
  calcDiff()
</script>

<style scoped>
  .tool-panel { padding: 8px 0; }
  .hint { font-size: 12px; color: #909399; margin: 6px 0 0; }
  .mt-2 { margin-top: 8px; }
  .mt-3 { margin-top: 12px; }
  .mb-2 { margin-bottom: 8px; }
  .w-full { width: 100%; }
  .label { font-size: 13px; color: #606266; display: block; margin-bottom: 4px; }
</style>
