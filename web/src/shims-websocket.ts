import Vue from 'vue'

declare module 'vue/types/vue' {
  interface Vue {
    $websocket: (path: string) => WebSocket;
  }
}
