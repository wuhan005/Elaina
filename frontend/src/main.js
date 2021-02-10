import Vue from 'vue'
import App from './App.vue'
import './registerServiceWorker'
import router from './router'
import AtComponents from 'at-ui'
import 'at-ui-style'
import utils from './utils'

Vue.config.productionTip = false
Vue.prototype.utils = utils
Vue.use(AtComponents)

new Vue({
    router,
    render: function (h) {
        return h(App)
    }
}).$mount('#app')
