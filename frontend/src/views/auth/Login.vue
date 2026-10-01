<template>
  <main class="login-page">
    <div class="ambient ambient-left" aria-hidden="true"></div>
    <div class="ambient ambient-right" aria-hidden="true"></div>

    <section class="login-panel" aria-label="登录">
      <div class="brand" aria-label="品牌标识">
        <span class="brand-mark">
          <img
            v-if="siteSettings.logoUrl"
            :src="siteSettings.logoUrl"
            :alt="`${siteSettings.siteName} logo`"
          />
          <span v-else>轻</span>
        </span>
        <span class="brand-name">{{ siteSettings.siteName }}</span>
      </div>

      <form class="login-form" novalidate @submit.prevent="handleLogin">
        <label class="field-label" for="username">账号</label>
        <div class="field-wrap" :class="{ 'has-error': errorMessage && !username }">
          <svg class="field-icon" viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="12" cy="8" r="3.25" />
            <path d="M5.5 19c.45-3.17 2.6-4.75 6.5-4.75s6.05 1.58 6.5 4.75" />
          </svg>
          <input
            id="username"
            v-model="username"
            name="username"
            type="text"
            autocomplete="username"
            placeholder="请输入账号"
            autofocus
          />
        </div>

        <label class="field-label password-label" for="password">密码</label>
        <div class="field-wrap" :class="{ 'has-error': errorMessage && !password }">
          <svg class="field-icon" viewBox="0 0 24 24" aria-hidden="true">
            <rect x="5.25" y="10.25" width="13.5" height="10" rx="2" />
            <path d="M8.25 10.25V7.5a3.75 3.75 0 0 1 7.5 0v2.75" />
          </svg>
          <input
            id="password"
            v-model="password"
            name="password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="current-password"
            placeholder="请输入密码"
          />
          <button
            class="password-toggle"
            type="button"
            :aria-label="showPassword ? '隐藏密码' : '显示密码'"
            @click="showPassword = !showPassword"
          >
            <svg v-if="showPassword" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M3.5 12s3.1-5 8.5-5 8.5 5 8.5 5-3.1 5-8.5 5-8.5-5-8.5-5Z" />
              <circle cx="12" cy="12" r="2.25" />
            </svg>
            <svg v-else viewBox="0 0 24 24" aria-hidden="true">
              <path d="m4 4 16 16M10.6 6.96A9.1 9.1 0 0 1 12 6.86c5.4 0 8.5 5.14 8.5 5.14a15.8 15.8 0 0 1-2.18 2.62M6.1 6.84C4.43 8.1 3.5 12 3.5 12s3.1 5.14 8.5 5.14c.94 0 1.8-.15 2.57-.4" />
            </svg>
          </button>
        </div>

        <p v-if="errorMessage" class="form-error" role="alert">{{ errorMessage }}</p>

        <button class="submit-button" type="submit" :disabled="isSubmitting">
          <span v-if="isSubmitting" class="loading-dot" aria-hidden="true"></span>
          {{ isSubmitting ? '登录中…' : '登录' }}
        </button>
      </form>

    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  useRoute,
  useRouter,
} from 'vue-router'
import { loadSiteSettings, siteSettings } from '../../stores/siteSettings'

const router = useRouter()
const route = useRoute()
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')

onMounted(() => {
  loadSiteSettings()
})

const handleLogin = async () => {
  errorMessage.value = ''

  if (!username.value.trim() || !password.value) {
    errorMessage.value = '请输入账号和密码'
    return
  }

  isSubmitting.value = true
  await new Promise((resolve) => window.setTimeout(resolve, 360))

  if (
    username.value.trim() !== siteSettings.admin.username ||
    password.value !== siteSettings.admin.password
  ) {
    errorMessage.value = '账号或密码不正确，请重试'
    isSubmitting.value = false
    return
  }

  window.sessionStorage.setItem('lightdocs-authenticated', 'true')
  const redirect = typeof route.query.redirect === 'string'
    && route.query.redirect.startsWith('/admin')
    ? route.query.redirect
    : '/admin'

  await router.push(redirect)
  isSubmitting.value = false
}
</script>

<style scoped>
.login-page {
  position: relative;
  display: grid;
  min-height: 100vh;
  place-items: center;
  overflow: hidden;
  padding: 32px 20px;
  background:
    radial-gradient(circle at 50% 0%, rgba(224, 233, 255, 0.82), transparent 48%),
    #f6f8fc;
}

.ambient {
  position: absolute;
  pointer-events: none;
  border-radius: 50%;
  filter: blur(1px);
}

.ambient-left {
  width: 340px;
  height: 340px;
  left: -170px;
  bottom: -145px;
  background: rgba(124, 92, 255, 0.06);
}

.ambient-right {
  width: 260px;
  height: 260px;
  top: -135px;
  right: -105px;
  background: rgba(71, 191, 255, 0.08);
}

.login-panel {
  position: relative;
  z-index: 1;
  width: min(100%, 410px);
  padding: 44px 42px 32px;
  border: 1px solid rgba(224, 229, 240, 0.9);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.93);
  box-shadow: 0 26px 70px rgba(55, 73, 121, 0.12), 0 2px 8px rgba(55, 73, 121, 0.04);
  backdrop-filter: blur(12px);
}

.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 11px;
}

.brand-mark {
  display: grid;
  width: 39px;
  height: 39px;
  place-items: center;
  overflow: hidden;
  border-radius: 11px;
  background: linear-gradient(145deg, #8258ff, #4b8cff);
  box-shadow: 0 8px 18px rgba(105, 86, 228, 0.25);
  color: #fff;
  font-size: 18px;
  font-weight: 700;
}

.brand-mark img {
  width: 25px;
  height: 25px;
  object-fit: contain;
}

.brand-name {
  color: #1d2842;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.login-form {
  margin-top: 36px;
}

.field-label {
  display: block;
  margin-bottom: 9px;
  color: #48536b;
  font-size: 13px;
  font-weight: 600;
}

.password-label {
  margin-top: 20px;
}

.field-wrap {
  position: relative;
  display: flex;
  align-items: center;
  height: 50px;
  border: 1px solid #e0e5ef;
  border-radius: 11px;
  background: #fbfcfe;
  transition: border-color 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

.field-wrap:focus-within {
  border-color: #6f74ed;
  background: #fff;
  box-shadow: 0 0 0 4px rgba(111, 116, 237, 0.1);
}

.field-wrap.has-error {
  border-color: #e47777;
}

.field-icon {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  margin: 0 12px 0 15px;
  fill: none;
  stroke: #a5aec0;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.65;
}

.field-wrap input {
  min-width: 0;
  height: 100%;
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  color: #26314a;
  font-size: 14px;
}

.field-wrap input::placeholder {
  color: #b0b8c9;
}

.password-toggle {
  display: grid;
  width: 42px;
  height: 100%;
  place-items: center;
  flex: 0 0 auto;
  border: 0;
  background: transparent;
  color: #a5aec0;
  cursor: pointer;
}

.password-toggle:hover {
  color: #6f74ed;
}

.password-toggle svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.55;
}

.form-error {
  margin: 12px 0 -2px;
  color: #d65f67;
  font-size: 12px;
}

.submit-button {
  display: flex;
  width: 100%;
  height: 50px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 27px;
  border: 0;
  border-radius: 11px;
  background: linear-gradient(100deg, #655ee8, #7a72ee);
  box-shadow: 0 11px 20px rgba(101, 94, 232, 0.2);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 160ms ease, box-shadow 160ms ease, opacity 160ms ease;
}

.submit-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 14px 25px rgba(101, 94, 232, 0.28);
}

.submit-button:active:not(:disabled) {
  transform: translateY(0);
}

.submit-button:disabled {
  cursor: wait;
  opacity: 0.75;
}

.loading-dot {
  width: 13px;
  height: 13px;
  border: 2px solid rgba(255, 255, 255, 0.45);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 700ms linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 480px) {
  .login-page {
    padding: 18px;
  }

  .login-panel {
    padding: 34px 24px 27px;
    border-radius: 18px;
  }
}
</style>
