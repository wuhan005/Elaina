import {createApp} from 'vue'

import App from './App.vue'
import TDesign from 'tdesign-vue-next';
import './theme.css'
import 'tdesign-vue-next/es/style/index.css';
import router from './route/index'

const app = createApp(App)
app.use(TDesign)
app.use(router)
app.mount('#app')
