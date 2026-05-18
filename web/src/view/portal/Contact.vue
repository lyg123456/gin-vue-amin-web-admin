<template>
  <div class="contact-page">
    <section class="panel">
      <h1 class="page-title">联系方式</h1>
      <p class="lead">欢迎通过电话或微信联系我们；也可填写下方信息，我们会尽快回电。</p>

      <h2 class="section-title">留资方式</h2>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px" class="lead-form">
        <el-form-item label="电话号码" prop="phone">
          <el-input v-model="form.phone" maxlength="32" placeholder="请输入您的联系电话" clearable />
        </el-form-item>
        <el-form-item label="留言备注" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="4"
            maxlength="2000"
            show-word-limit
            placeholder="可填写需求说明、方便联系的时间等"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="onSubmit">提交</el-button>
        </el-form-item>
      </el-form>

      <h2 class="section-title">加微信</h2>
      <p class="wechat-hint">微信昵称：<strong>清风</strong> · 请扫描下方二维码添加好友</p>
      <div class="qr-wrap">
        <img class="qr-img" src="/portal/wechat-qingfeng.png" alt="微信二维码 — 清风" />
      </div>
      <p class="phone-line">电话：<a href="tel:19225501831">19225501831</a></p>
    </section>
  </div>
</template>

<script setup>
  import { ref, reactive } from 'vue'
  import { ElMessage } from 'element-plus'
  import { submitPortalContactLead } from '@/api/publicPortalContact'

  defineOptions({
    name: 'PortalContact'
  })

  const formRef = ref()
  const submitting = ref(false)
  const form = reactive({
    phone: '',
    remark: ''
  })

  const rules = {
    phone: [{ required: true, message: '请输入电话号码', trigger: 'blur' }]
  }

  const onSubmit = async () => {
    if (!formRef.value) return
    try {
      await formRef.value.validate()
    } catch {
      return
    }
    submitting.value = true
    try {
      const res = await submitPortalContactLead({
        phone: form.phone.trim(),
        remark: form.remark.trim()
      })
      if (res.code === 0) {
        ElMessage.success(res.msg || '提交成功')
        form.phone = ''
        form.remark = ''
        formRef.value.resetFields()
      }
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .contact-page {
    max-width: 640px;
    margin: 0 auto;
  }

  .panel {
    background: var(--portal-panel-bg, #fff);
    border-radius: var(--portal-radius, 12px);
    padding: 24px 22px 28px;
    box-sizing: border-box;
  }

  .page-title {
    margin: 0 0 10px;
    font-size: 1.35rem;
    font-weight: 700;
    color: #1a1a1a;
  }

  .lead {
    margin: 0 0 28px;
    font-size: 0.9rem;
    color: var(--portal-text-secondary, #6b7280);
    line-height: 1.55;
  }

  .section-title {
    margin: 0 0 14px;
    font-size: 1.05rem;
    font-weight: 700;
    color: #1a1a1a;
    padding-top: 8px;
    border-top: 1px solid var(--portal-hairline, #f3f4f6);
  }

  .section-title:first-of-type {
    border-top: none;
    padding-top: 0;
  }

  .lead-form {
    margin-bottom: 8px;
  }

  .wechat-hint {
    margin: 0 0 14px;
    font-size: 0.9rem;
    color: var(--portal-text-body, #4b5563);
  }

  .qr-wrap {
    display: flex;
    justify-content: center;
    margin-bottom: 16px;
  }

  .qr-img {
    width: 100%;
    max-width: 260px;
    height: auto;
    border-radius: 8px;
    box-shadow: 0 4px 16px rgba(15, 23, 42, 0.08);
  }

  .phone-line {
    margin: 0;
    font-size: 0.9rem;
    color: var(--portal-text-body, #4b5563);
  }

  .phone-line a {
    color: var(--portal-link, #2563eb);
    text-decoration: none;
  }

  .phone-line a:hover {
    text-decoration: underline;
  }
</style>
