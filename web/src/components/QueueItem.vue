<template>
  <v-list-item :two-line="downloading || !!error" class="mx-0 px-0 my-0 py-0" dense>
    <v-list-item-action class="mr-2 my-0" v-if="!$vuetify.breakpoint.xs">
      <!-- vuedraggable is a big screw you to all the benefits of not tightly coupling dom and logic -->
      <v-icon class="drag">mdi-drag-vertical</v-icon>
    </v-list-item-action>
    <v-list-item-action class="mr-4 my-0 py-0">
      <v-btn depressed small color="deep-orange accent-1" @click="deleteSong"><v-icon>mdi-delete</v-icon></v-btn>
    </v-list-item-action>
    <v-list-item-content class="my-0 py-0">
      <v-list-item-title class="my-0 py-0"><span>{{ title }}</span></v-list-item-title>
      <v-list-item-subtitle v-if="downloading || !!error">
        <ProgressBar :error="error" :progress="progress"/>
      </v-list-item-subtitle>
    </v-list-item-content>
    <v-list-item-action class="my-0 py-0 ml-8">
      <v-btn depressed small color="deep-orange accent-1" :disabled="disabledown" @click="down"><v-icon>mdi-arrow-down-bold</v-icon></v-btn>
    </v-list-item-action>
    <v-list-item-action class="my-0 py-0">
      <v-btn depressed small color="deep-orange accent-1" :disabled="disableup" @click="up"><v-icon>mdi-arrow-up-bold</v-icon></v-btn>
    </v-list-item-action>
  </v-list-item>
</template>

<style scoped>
/*
 * font-size doesn't work on v-list-item-title and using the text classes causes text cutoff lmao
 */
span {
  font-size: 1.4em;
}

.drag {
  cursor: grab;
}

.drag:active {
  cursor: grabbing;
}
</style>

<script lang="ts">
import Vue from 'vue'
import ProgressBar from './ProgressBar.vue'

export default Vue.extend({
  name: 'QueueItem',

  components: {
    ProgressBar
  },

  methods: {
    deleteSong () {
      this.$http.delete(`queue/${this.index}`)
    },
    down () {
      this.$http.put(`queue/${this.index}`, new URLSearchParams({
        move: 'down'
      }))
    },
    up () {
      this.$http.put(`queue/${this.index}`, new URLSearchParams({
        move: 'up'
      }))
    }
  },

  props: ['disabledown', 'disableup', 'downloading', 'error', 'index', 'progress', 'title']
})
</script>
