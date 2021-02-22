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
            Dashboard
          </b-navbar-item>
          <b-navbar-item tag="router-link" :to="{ path: '/template' }">
            Template
          </b-navbar-item>
          <b-navbar-item tag="router-link" :to="{ path: '/sandbox' }">
            Sandbox
          </b-navbar-item>
          <b-navbar-item @click="logout">
            Logout
          </b-navbar-item>
        </template>
      </b-navbar>
      <div class="container-fluid">
        <router-view/>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  mounted() {
    this.utils.GET('/m/status').catch(err => {
      if (localStorage.getItem('login') !== null) {
        this.logout()
      }
    })
  },

  methods: {
    logout() {
      this.utils.POST('/m/logout')
      localStorage.removeItem('login')
      location.reload()
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