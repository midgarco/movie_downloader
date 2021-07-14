import 'core-js/stable'
import 'regenerator-runtime/runtime'
import Vue from 'vue'
import App from './App.vue'

Vue.config.productionTip = false
Vue.config.devtools = true
Vue.config.errorHandler = function(err, vm, info) {
  console.log(`Error: ${err.toString()}\nInfo: ${info}`);
}

import * as Wails from '@wailsapp/runtime'

Wails.Init(() => {
  new Vue({
    render: (h) => h(App),
  }).$mount('#app')
})