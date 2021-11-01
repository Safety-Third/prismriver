import Vue from 'vue'

const websocketBase = process.env.NODE_ENV === 'production' ? process.env.VUE_APP_WS_URL || `${window.location.protocol ===
  'https:' ? 'wss:' : 'ws:'}//${window.location.hostname}` : 'ws://localhost:8000'

Vue.prototype.$websocket = (path: string) => {
  return new WebSocket(`${websocketBase}/${path}`)
}
