import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import './assets/styles/global.css'
import './assets/styles/theme.css'
import './assets/styles/element-override.css'
import './assets/styles/list-page.css'
import './assets/styles/detail-page.css'
import './assets/styles/form-dialog.css'

import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user'
import { registerUnauthorizedClearHandler } from './utils/request'

const app = createApp(App)

app.use(createPinia())
registerUnauthorizedClearHandler(() => {
  useUserStore().clearUserData()
})
app.use(router)
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')
