<template>
  <div id="app">
    <div class="container">
      <at-menu mode="horizontal" :active-name="$route.name" style="width: 100%" @on-select="go">
        <at-menu-item name="dashboard"><i class="icon icon-bar-chart-2"></i>仪表盘</at-menu-item>
        <at-menu-item name="template"><i class="icon icon-copy"></i>运行模板</at-menu-item>
      </at-menu>
      <div class="container-fluid">
        <router-view style="margin-top: 10px"/>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  methods: {
    go(route) {
      this.$router.push({
        name: route
      })
    }
  }
}
</script>

<style>
#app {
  display: flex;
  height: 100%;
  justify-content: center;
}
</style>