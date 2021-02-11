<template>
  <div id="app">
    <div class="container">
      <b-navbar>
        <template #brand>
          <b-navbar-item tag="router-link" :to="{ path: '/' }">
            <span>Elaina</span>
          </b-navbar-item>
        </template>
        <template #start>
          <b-navbar-item tag="router-link" :to="{ path: '/' }">
            仪表盘
          </b-navbar-item>
          <b-navbar-item tag="router-link" :to="{ path: '/template' }">
            运行模板
          </b-navbar-item>
          <b-navbar-item tag="router-link" :to="{ path: '/sandbox' }">
            沙箱
          </b-navbar-item>
        </template>
      </b-navbar>
      <div class="container-fluid">
        <router-view/>
      </div>
    </div>
  </div>
</template>

<style>
#app {
  display: flex;
  height: 100%;
  justify-content: center;
}
</style>