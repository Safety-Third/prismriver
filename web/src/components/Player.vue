<template>
  <v-card outlined>
    <v-card-title class="deep-orange py-2 white--text">Currently Playing</v-card-title>
    <v-divider/>
    <v-card-text>
      <h3 class="black--text font-weight-medium text-h6 text-truncate">{{ title }}</h3>
      <ProgressBar v-if="item && (item.downloading || !!item.error)" :error="item.error" :progress="item.progress"/>
    </v-card-text>
    <v-card-actions class="mb-0 pb-0">
      <v-slider inverse-label :label="progress" v-model.number="currentTime" :max="totalTime" @end="seek" @start="seeking = true"/>
    </v-card-actions>
    <v-card-actions class="mt-0 pt-0">
      <v-btn class="mx-2" depressed small color="deep-orange accent-1" @click="volDown"><v-icon>mdi-volume-minus</v-icon></v-btn>
      <!-- to future maintainers: please let eman live on -->
      <h4>{{ volume }}</h4>
      <v-btn class="mx-2" depressed small color="deep-orange accent-1" @click="volUp"><v-icon>mdi-volume-plus</v-icon></v-btn>
      <v-spacer/>
      <v-btn depressed small class="text-none text-h6" color="deep-orange accent-1" @click="skip"><v-icon>mdi-skip-forward</v-icon> Skip</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import Vue from 'vue'
import ProgressBar from './ProgressBar.vue'

export default Vue.extend({
  name: 'Player',

  components: {
    ProgressBar
  },

  computed: {
    progress (): string {
      return `${this.$parseTime(this.currentTime / 1000)} / ${this.$parseTime(this.totalTime / 1000)}`
    },
    title () {
      if (this.item) {
        return this.item.media.Title
      } else {
        return 'Nothing Currently Playing'
      }
    }
  },

  data: () => ({
    currentTime: 0,
    seeking: false,
    socket: null as WebSocket | null,
    state: 0,
    totalTime: 1,
    volume: 100,
    ws: 0
  }),

  methods: {
    connectWS () {
      this.socket = this.$websocket('ws/player')

      this.socket.addEventListener('close', () => {
        this.ws = 0
      })
      this.socket.addEventListener('error', () => {
        this.ws = 2
      })
      this.socket.addEventListener('message', (event: MessageEvent) => {
        this.ws = 1
        const data = JSON.parse(event.data)
        this.currentTime = data.CurrentTime
        this.totalTime = data.TotalTime
        this.state = data.State
        this.volume = data.Volume

        // more vuetify jankiness pyon~
        if (!this.totalTime) {
          this.totalTime = 1
        }
      })
    },
    seek (time: number) {
      this.seeking = false
      if (this.state !== 1) {
        return
      }
      this.$http.put('player', new URLSearchParams({
        seek: time.toString()
      }))
    },
    skip () {
      this.$http.delete('queue/0')
    },
    volDown () {
      this.$http.put('player', new URLSearchParams({
        volume: 'down'
      }))
    },
    volUp () {
      this.$http.put('player', new URLSearchParams({
        volume: 'up'
      }))
    }
  },

  mounted () {
    this.connectWS()
    setInterval(() => {
      if (this.state === 1 && !this.seeking) {
        this.currentTime += 1000
      }
    }, 1000)

    setInterval(() => {
      if (this.ws === 0) {
        this.connectWS()
      }
    }, 5000)
  },

  props: ['item'],

  watch: {
    ws (state) {
      this.$emit('update:ws', state)
    }
  }
})
</script>
