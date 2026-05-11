<template>
  <section v-if="slides.length" class="portal-hero" aria-label="首页轮播">
    <el-carousel
      class="hero-carousel"
      :height="carouselHeight"
      :interval="6000"
      indicator-position="outside"
      arrow="hover"
    >
      <el-carousel-item v-for="(item, idx) in slides" :key="idx">
        <div
          class="slide"
          :style="{ backgroundImage: `url(${resolveImg(item.image)})` }"
        >
          <div class="slide-overlay" />
          <div class="slide-content">
            <h2 class="slide-title">{{ item.title }}</h2>
            <p v-if="item.subtitle" class="slide-sub">{{ item.subtitle }}</p>
            <el-button
              v-if="item.linkText && item.linkUrl"
              type="primary"
              size="large"
              round
              class="slide-btn"
              @click="onAction(item.linkUrl)"
            >
              {{ item.linkText }}
            </el-button>
          </div>
        </div>
      </el-carousel-item>
    </el-carousel>
  </section>
</template>

<script setup>
  import { computed, onMounted, onUnmounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { getPublicHomeCarousel } from '@/api/publicPortal'
  import { getUrl } from '@/utils/image'

  const router = useRouter()
  const slides = ref([])
  const winW = ref(typeof window !== 'undefined' ? window.innerWidth : 960)

  const carouselHeight = computed(() => (winW.value <= 640 ? '260px' : '360px'))

  const onResize = () => {
    winW.value = window.innerWidth
  }

  const resolveImg = (u) => {
    if (!u) return ''
    const s = String(u).trim()
    if (/^https?:\/\//i.test(s)) return s
    return getUrl(s)
  }

  const onAction = (url) => {
    if (!url) return
    if (/^https?:\/\//i.test(url)) {
      window.open(url, '_blank', 'noopener,noreferrer')
      return
    }
    const path = url.startsWith('/') ? url : `/${url}`
    router.push(path)
  }

  onMounted(async () => {
    window.addEventListener('resize', onResize)
    try {
      const res = await getPublicHomeCarousel()
      if (res.code === 0) {
        slides.value = Array.isArray(res.data?.list) ? res.data.list : []
      }
    } catch {
      slides.value = []
    }
  })

  onUnmounted(() => {
    window.removeEventListener('resize', onResize)
  })
</script>

<style scoped>
  .portal-hero {
    width: 100%;
    max-width: 100%;
    margin: 0 auto;
    border-radius: var(--portal-radius, 12px);
    overflow: hidden;
    background: #111;
  }

  .hero-carousel {
    width: 100%;
  }

  .slide {
    position: relative;
    height: 100%;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
  }

  .slide-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, rgba(0, 0, 0, 0.35) 0%, rgba(0, 0, 0, 0.55) 100%);
  }

  .slide-content {
    position: relative;
    z-index: 1;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 24px 20px;
    box-sizing: border-box;
  }

  .slide-title {
    margin: 0 0 12px;
    max-width: 920px;
    font-size: clamp(1.25rem, 3.5vw, 1.85rem);
    font-weight: 700;
    line-height: 1.35;
    color: #ffffff;
    text-shadow: 0 2px 12px rgba(0, 0, 0, 0.35);
  }

  .slide-sub {
    margin: 0 0 22px;
    max-width: 720px;
    font-size: clamp(0.85rem, 2vw, 1rem);
    line-height: 1.6;
    color: rgba(255, 255, 255, 0.92);
    text-shadow: 0 1px 8px rgba(0, 0, 0, 0.35);
  }

  .slide-btn {
    min-width: 120px;
    font-weight: 600;
  }

  /* 条状指示器，贴近常见门户轮播 */
  .hero-carousel :deep(.el-carousel__indicators--outside) {
    margin-top: 10px;
    padding-bottom: 8px;
  }
  .hero-carousel :deep(.el-carousel__indicator--horizontal) {
    padding: 0 4px;
  }
  .hero-carousel :deep(.el-carousel__indicator--horizontal .el-carousel__button) {
    width: 28px;
    height: 4px;
    border-radius: 2px;
    background-color: rgba(255, 255, 255, 0.45);
  }
  .hero-carousel :deep(.el-carousel__indicator--horizontal.is-active .el-carousel__button) {
    background-color: #3b82f6;
  }
</style>
