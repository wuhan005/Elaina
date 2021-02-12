<template>
  <div>
    <section class="hero">
      <div class="hero-body">
        <p class="title">登录</p>
        <p class="subtitle">请输入密码以登录 Elaina</p>
      </div>
    </section>

    <form action="">
      <div class="modal-card">
        <section class="modal-card-body">
          <b-field label="密码">
            <b-input type="password" v-model="password" password-reveal required>
            </b-input>
          </b-field>

        </section>
        <footer class="modal-card-foot">
          <b-button label="登录" type="is-primary" @click="login"/>
        </footer>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: "Login",

  data() {
    return {
      password: '',
    }
  },

  methods: {
    login() {
      this.utils.POST('/m/login', {
        password: this.password
      }).then(res => {
        this.$buefy.toast.open({message: res, type: 'is-success'})
        localStorage.setItem('login', true);
        this.$router.push('/')
      }).catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    }
  }
}
</script>

<style scoped>

</style>