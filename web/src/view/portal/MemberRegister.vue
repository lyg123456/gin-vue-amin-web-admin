<template>
  <div class="member-register">
    <section class="panel">
      <h1>会员注册</h1>
      <p class="lead">
        当前演示工程未单独实现「前台会员注册」接口。实际项目中可自建会员表与注册 API，或使用短信/三方登录等方案。
      </p>
      <ul>
        <li>若账号由<strong>管理员在后台创建</strong>，请直接使用「会员登录」。</li>
        <li>开源框架自带的 <code>/user/admin_register</code> 为<strong>管理员注册</strong>能力，不适合直接暴露给公网会员。</li>
      </ul>
      <div class="actions">
        <el-button type="primary" @click="$router.push({ name: 'MemberEntryLogin' })">去会员登录</el-button>
        <el-button @click="$router.push({ name: 'WebMember' })">返回会员中心</el-button>
      </div>
    </section>
  </div>
</template>

<script setup>
</script>

<style scoped>
  .member-register {
    max-width: 640px;
  }
  .panel {
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    border: none;
    padding: 24px;
  }
  h1 {
    margin: 0 0 12px;
    font-size: 1.5rem;
  }
  .lead {
    margin: 0 0 16px;
    color: var(--portal-text-body, #4b5563);
    line-height: 1.65;
    font-size: 0.95rem;
  }
  ul {
    margin: 0 0 24px;
    padding-left: 1.2rem;
    color: var(--portal-text-body, #4b5563);
    line-height: 1.8;
  }
  code {
    font-size: 0.85rem;
    background: #f3f4f6;
    padding: 2px 6px;
    border-radius: 4px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
</style>
