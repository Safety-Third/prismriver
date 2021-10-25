<template>
  <v-app>
    <v-app-bar app dense flat color="primary">
      <v-toolbar-title class="white--text" style="width: 100%;">2GroovE <v-icon class="white--text">mdi-music</v-icon></v-toolbar-title>
      <v-spacer/>
      <v-btn depressed small class="text-none" color="secondary black--text text-subtitle-1" @click="beQuiet"><v-icon>mdi-alert-circle</v-icon>Be Quiet!</v-btn>
    </v-app-bar>

    <v-main>
      <v-container>
        <v-row>
          <v-col cols="2" v-if="!$vuetify.breakpoint.xs"/>
          <v-col :cols="$vuetify.breakpoint.xs ? 12 : 8">
            <v-alert v-if="state === 0" type="error">The player has disconnected from the server. Attempting to reconnect...</v-alert>
            <v-alert v-if="state === 2" type="warning">The page has encountered errors. The current queue may be outdated.</v-alert>
            <URLForm/>
            <br>
            <SearchForm/>
            <br>
            <Player :item="queue[0]" @update:ws="playerWS = $event"/>
            <br>
            <v-card outlined>
              <v-card-title class="mb-0 py-2">
                <span>Current Queue</span>
                <v-spacer/>
                <v-btn depressed small color="deep-orange accent-1" @click="shuffle">
                  <v-icon>mdi-shuffle</v-icon>
                </v-btn>
              </v-card-title>
              <v-divider/>
              <v-card-text class="mb-0 pb-0 text-center" v-if="queue.length < 2">
                <h3>Queue empty. Add some music!</h3>
              </v-card-text>
              <v-card-text class="mb-0 pb-0" v-else>
                <!-- vuetify's built-in animation tools are jank -->
                <transition-group name="queue">
                  <QueueItem class="queue-item" v-for="(item, i) in queue.slice(1)" :key="item.media.ID"
                    :disabledown="i === queue.length - 2" :disableup="i === 0" :downloading="item.downloading"
                    :error="item.error" :index="i + 1" :progress="item.progress" :title="item.media.Title"/>
                </transition-group>
              </v-card-text>
              <v-card-actions/>
            </v-card>
          </v-col>
          <v-col cols="2" v-if="!$vuetify.breakpoint.xs"/>
        </v-row>
      </v-container>
    </v-main>

    <v-footer color="primary white--text">
      <v-col cols="auto">2GroovE the website; Made by 2E residents one time they were bored.</v-col>
      <v-spacer/>
      <v-col class="text-right"><a class="white--text" href="https://next2e.github.io"><v-icon color="white">mdi-earth</v-icon> 2E Website</a></v-col>
    </v-footer>
  </v-app>
</template>

<style scoped>
.queue-enter, .queue-leave-to {
  opacity: 0;
  transform: translateX(32px);
}

/*
 * can't use QueueItem over class selector because this is also jank lmao
 * also can't use move classes for some god forsaken reason
 */
.queue-item {
  transition: all 1s;
}

.queue-leave-active {
  position: absolute;
}

.v-card {
  border-color: darkgray;
}
</style>

<script lang="ts">
import Vue from 'vue'
import Player from './components/Player.vue'
import QueueItem from './components/QueueItem.vue'
import SearchForm from './components/SearchForm.vue'
import URLForm from './components/URLForm.vue'

export default Vue.extend({
  name: 'App',

  components: {
    Player,
    QueueItem,
    SearchForm,
    URLForm
  },

  computed: {
    state () {
      return Math.max(this.playerWS, this.queueWS)
    }
  },

  data: () => ({
    playerWS: 0,
    queue: [],
    queueWS: 0,
    results: [],
    socket: null as WebSocket | null
  }),

  methods: {
    beQuiet () {
      this.$http.put('player', new URLSearchParams({
        quiet: 'true'
      }))
    },
    connectWS () {
      this.socket = this.$websocket('ws/queue')

      this.socket.addEventListener('close', () => {
        this.queueWS = 0
      })
      this.socket.addEventListener('error', () => {
        this.queueWS = 2
      })
      this.socket.addEventListener('message', (event: MessageEvent) => {
        this.queueWS = 1
        const queue = JSON.parse(event.data)
        this.queue = queue
      })
    },
    shuffle () {
      this.$http.put('player', new URLSearchParams({
        shuffle: 'true'
      }))
    }
  },

  async mounted () {
    try {
      const config = (await this.$http.get('config.json')).data
      this.$http.defaults.baseURL = config.API_URL
    } catch (e) {}

    this.connectWS()

    setInterval(() => {
      if (this.queueWS === 0) {
        this.connectWS()
      }
    }, 5000)
  }
})
</script>
