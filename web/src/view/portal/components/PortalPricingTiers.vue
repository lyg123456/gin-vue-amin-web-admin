<template>
  <section class="portal-pricing">
    <h2 class="portal-pricing__title">服务套餐</h2>

    <div class="pricing-grid">
      <div class="pricing-row">
        <div class="pricing-cell pricing-cell--left">
          <div class="text-block text-right">
            <h3 class="title">99元 · 网站建设</h3>
            <p class="desc">官网建设3套模板，域名备案，服务器购买服务</p>
          </div>
          <div class="icon-circle icon-circle--yellow">
            <el-icon :size="28"><View /></el-icon>
          </div>
        </div>
        <div class="pricing-cell pricing-cell--right">
          <div class="icon-circle icon-circle--orange">
            <el-icon :size="28"><Opportunity /></el-icon>
          </div>
          <div class="text-block text-left">
            <h3 class="title">199元 · 小程序搭建</h3>
            <p class="desc">小程序搭建，后台管理系统</p>
          </div>
        </div>
      </div>

      <div class="pricing-row">
        <div class="pricing-cell pricing-cell--left">
          <div class="text-block text-right">
            <h3 class="title">499元 · APP开发</h3>
            <p class="desc">APP开发发布应用商店，后台管理系统</p>
          </div>
          <div class="icon-circle icon-circle--green">
            <el-icon :size="28"><List /></el-icon>
          </div>
        </div>
        <div class="pricing-cell pricing-cell--right">
          <div class="icon-circle icon-circle--blue">
            <el-icon :size="28"><Cellphone /></el-icon>
          </div>
          <div class="text-block text-left">
            <h3 class="title">899元 · AI自动化</h3>
            <p class="desc">AI自动发布文章；自动发布AI短视频-文生图；AI图生视频</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
  import { View, Opportunity, List, Cellphone } from '@element-plus/icons-vue'

  defineOptions({
    name: 'PortalPricingTiers'
  })
</script>

<style lang="scss" scoped>
  .portal-pricing {
    background: var(--portal-panel-bg, #ffffff);
    border-radius: var(--portal-radius, 12px);
    padding: 24px 20px 28px;
    box-sizing: border-box;
  }

  .portal-pricing__title {
    margin: 0 0 28px;
    text-align: center;
    font-size: 1.05rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    color: #1a1a1a;
  }

  .pricing-grid {
    display: flex;
    flex-direction: column;
    gap: 2.5rem;
    max-width: 80rem;
    margin: 0 auto;
  }

  .pricing-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem 2rem;
    align-items: center;

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }

  .pricing-cell {
    display: flex;
    align-items: center;
    gap: 1.25rem;
    min-height: 5.5rem;
  }

  .pricing-cell--left {
    justify-content: flex-end;
  }

  .pricing-cell--right {
    justify-content: flex-start;
  }

  @media (max-width: 768px) {
    .pricing-cell--left,
    .pricing-cell--right {
      justify-content: center;
    }

    .text-block.text-right {
      text-align: left;
      order: 2;
    }

    .pricing-cell--left .icon-circle {
      order: 1;
    }
  }

  .text-block {
    max-width: 16rem;
    flex-shrink: 1;
  }

  .title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    line-height: 1.4;
    color: #334155;
  }

  .desc {
    margin: 0.35rem 0 0;
    font-size: 0.8125rem;
    line-height: 1.55;
    color: #64748b;
  }

  .icon-circle {
    flex-shrink: 0;
    width: 4.5rem;
    height: 4.5rem;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    box-shadow: 0 4px 14px rgba(15, 23, 42, 0.12);
  }

  .icon-circle--yellow {
    background: #f5c542;
  }

  .icon-circle--orange {
    background: #f59e0b;
  }

  .icon-circle--green {
    background: #22c55e;
  }

  .icon-circle--blue {
    background: #38bdf8;
  }
</style>
