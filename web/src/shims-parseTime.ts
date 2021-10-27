import Vue from 'vue'

declare module 'vue/types/vue' {
  interface Vue {
    $parseTime: (time: number, recur?: boolean) => string;
  }
}
