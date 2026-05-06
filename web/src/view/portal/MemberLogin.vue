<template>
  <div class="member-login">
    <div class="card">
      <h1>会员登录</h1>
      <p class="sub">登录站点会员账号，与后台管理员入口无关。</p>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @keyup.enter="submit">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" size="large" autocomplete="username" placeholder="用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            size="large"
            show-password
            autocomplete="current-password"
            placeholder="密码"
          />
        </el-form-item>
        <el-form-item v-if="form.openCaptcha" label="验证码" prop="captcha">
          <div class="captcha-row">
            <el-input v-model="form.captcha" size="large" placeholder="验证码" />
            <div class="pic" @click="refreshCaptcha">
              <img v-if="picPath" :src="picPath" alt="验证码" />
            </div>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="w-full" :loading="submitting" @click="submit">登 录</el-button>
        </el-form-item>
      </el-form>

      <div class="footer-links">
        <router-link :to="{ name: 'WebMember' }">返回会员中心</router-link>
        <span class="sep">·</span>
        <router-link :to="{ name: 'WebMemberRegister' }">没有账号？</router-link>
        <span class="sep">·</span>
        <router-link :to="{ name: 'WebHome' }">回首页</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { reactive, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { ElMessage } from 'element-plus'
  import { captcha } from '@/api/user'
  import { useUserStore } from '@/pinia/modules/user'

  const route = useRoute()
  const router = useRouter()
  const userStore = useUserStore()

  const formRef = ref(null)
  const picPath = ref('')
  const captchaLen = ref(6)
  const submitting = ref(false)

  const form = reactive({
    username: '',
    password: '',
    captcha: '',
    captchaId: '',
    openCaptcha: false
  })

  const checkCaptcha = (rule, value, callback) => {
    if (!form.openCaptcha) return callback()
    const v = (value || '').replace(/\s+/g, '')
    if (!v) return callback(new Error('请输入验证码'))
    if (!/^\d+$/.test(v)) return callback(new Error('验证码为数字'))
    if (v.length < captchaLen.value) return callback(new Error(`至少 ${captchaLen.value} 位`))
    if (v !== value) form.captcha = v
    callback()
  }

  const rules = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
    captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
  }

  const refreshCaptcha = async () => {
    const ele = await captcha()
    captchaLen.value = Number(ele.data?.captchaLength) || 0
    picPath.value = ele.data?.picPath
    form.captchaId = ele.data?.captchaId
    form.openCaptcha = ele.data?.openCaptcha
  }
  refreshCaptcha()

  const submit = () => {
    formRef.value?.validate(async (ok) => {
      if (!ok) return
      submitting.value = true
      try {
        const flag = await userStore.MemberLoginIn({ ...form })
        if (!flag) {
          await refreshCaptcha()
          return
        }
        ElMessage.success('登录成功')
        const redir = route.query.redirect
        const safe =
          typeof redir === 'string' &&
          redir.startsWith('/') &&
          !redir.startsWith('//') &&
          !redir.startsWith('/layout')
            ? redir
            : '/member'
        await router.replace(safe)
      } finally {
        submitting.value = false
      }
    })
  }
</script>

<style scoped>
  .member-login {
    display: flex;
    justify-content: center;
    padding: 12px 0 32px;
  }
  .card {
    width: 100%;
    max-width: 420px;
    background: #fff;
    border: 1px solid #e8eaed;
    border-radius: 12px;
    padding: 28px 24px 20px;
  }
  h1 {
    margin: 0 0 8px;
    font-size: 1.5rem;
  }
  .sub {
    margin: 0 0 24px;
    font-size: 0.9rem;
    color: #6b7280;
  }
  .captcha-row {
    display: flex;
    gap: 12px;
    width: 100%;
    align-items: center;
  }
  .pic {
    width: 120px;
    height: 40px;
    flex-shrink: 0;
    background: #e5e7eb;
    border-radius: 6px;
    overflow: hidden;
    cursor: pointer;
  }
  .pic img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .w-full {
    width: 100%;
  }
  .footer-links {
    text-align: center;
    font-size: 0.85rem;
    color: #6b7280;
    padding-top: 8px;
  }
  .footer-links a {
    color: #2563eb;
    text-decoration: none;
  }
  .footer-links a:hover {
    text-decoration: underline;
  }
  .sep {
    margin: 0 6px;
    color: #d1d5db;
  }
</style>
