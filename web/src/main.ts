import {createApp} from 'vue'

import App from './App.vue'
import TDesign from 'tdesign-vue-next';
import './theme.css'
import 'tdesign-vue-next/es/style/index.css';
import store from './store'
import router from './route/index'
import './api/interceptor'
import '@/style/index.less'

const app = createApp(App)
app.use(TDesign)
app.use(store)
app.use(router)
app.mount('#app')
