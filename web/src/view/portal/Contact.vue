<template>
  <div class="contact-page">
    <section class="portal-panel">
      <div class="page-title-bar">
        <h1 class="page-h1">联系方式</h1>
        <p class="page-desc">欢迎通过电话或微信联系我们；也可填写下方信息，我们会尽快回电。</p>
      </div>

      <div class="panel-body">
        <h2 class="section-title">留资方式</h2>
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="lead-form">
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
            <el-button type="primary" size="default" :loading="submitting" @click="onSubmit">提交</el-button>
          </el-form-item>
        </el-form>

        <h2 class="section-title">加微信</h2>
        <p class="wechat-hint">微信昵称：<strong>清风</strong> · 请扫描下方二维码添加好友</p>
        <div class="qr-wrap">
          <img class="qr-img" src="/portal/wechat-qingfeng.png" alt="微信二维码 — 清风" />
        </div>
        <p class="phone-line">电话：<a href="tel:19225501831">19225501831</a></p>
      </div>
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

<style lang="scss" scoped>
  .contact-page {
    padding-bottom: 16px;
  }

  .portal-panel {
    background: var(--portal-panel-bg, #fff);
    border-radius: var(--portal-radius, 4px);
    border: 1px solid #e0e0e0;
    padding: 0;
    overflow: hidden;
    box-sizing: border-box;
  }

  .page-title-bar {
    padding: 14px 16px 12px;
    border-bottom: 1px solid #eee;
    background: #fafafa;
    text-align: center;
  }

  .page-h1 {
    margin: 0;
    font-size: var(--portal-font-title, 18px);
    font-weight: 700;
    color: #333;
    line-height: 1.4;
  }

  .page-desc {
    margin: 6px 0 0;
    font-size: 14px;
    color: #888;
    line-height: 1.6;
  }

  .panel-body {
    padding: 20px 16px 28px;
    max-width: 520px;
    margin: 0 auto;
    text-align: center;
  }

  .section-title {
    margin: 0 0 14px;
    font-size: 16px;
    font-weight: 700;
    color: #333;
    padding-top: 16px;
    border-top: 1px solid #eee;
    text-align: center;
  }

  .section-title:first-of-type {
    border-top: none;
    padding-top: 0;
  }

  .lead-form {
    margin: 0 auto 8px;
    max-width: 400px;
    text-align: left;

    :deep(.el-form-item__label) {
      font-size: 14px;
      color: #333;
      justify-content: center;
      padding-bottom: 4px;
    }

    :deep(.el-form-item:last-child) {
      margin-bottom: 0;

      .el-form-item__content {
        justify-content: center;
      }
    }

    :deep(.el-input__inner),
    :deep(.el-textarea__inner) {
      font-size: 14px;
    }

    :deep(.el-button) {
      font-size: 14px;
      min-width: 120px;
    }
  }

  .wechat-hint {
    margin: 0 0 14px;
    font-size: 14px;
    color: var(--portal-text-body, #444);
    line-height: 1.6;

    strong {
      color: #333;
      font-weight: 600;
    }
  }

  .qr-wrap {
    display: flex;
    justify-content: center;
    margin-bottom: 16px;
  }

  .qr-img {
    width: 100%;
    max-width: 220px;
    height: auto;
    border-radius: 4px;
    border: 1px solid #e0e0e0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }

  .phone-line {
    margin: 0;
    font-size: 14px;
    color: var(--portal-text-body, #444);
  }

  .phone-line a {
    color: var(--portal-brand, #1a73e8);
    text-decoration: none;
    font-weight: 500;
  }

  .phone-line a:hover {
    text-decoration: underline;
    color: var(--portal-brand-dark, #1557b0);
  }

  @media (max-width: 640px) {
    .panel-body {
      padding: 16px 12px 24px;
    }

    .lead-form {
      max-width: 100%;
    }
  }
</style>
