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
    <v-list-item-action class="my-0 py-0 ml-4" v-if="$vuetify.breakpoint.xs">
      <v-btn depressed small color="deep-orange accent-1" :disabled="disabledown" @click="move('down')"><v-icon>mdi-arrow-down-bold</v-icon></v-btn>
    </v-list-item-action>
    <v-list-item-action class="my-0 py-0" v-if="$vuetify.breakpoint.xs">
      <v-btn depressed small color="deep-orange accent-1" :disabled="disableup" @click="move('up')"><v-icon>mdi-arrow-up-bold</v-icon></v-btn>
    </v-list-item-action>
    <v-list-item-action class="ml-4 my-0 py-0" v-else>
      <v-menu :disabled="disableup && disabledown">
        <template v-slot:activator="{ on, attrs }">
          <v-btn plain small text :ripple="false" v-bind="attrs" v-on="on" :disabled="disableup && disabledown"><v-icon>mdi-dots-vertical</v-icon></v-btn>
        </template>
        <v-list>
          <v-list-item v-if="!disableup" dense @click="move('top')"><v-list-item-title><v-icon>mdi-arrow-collapse-up</v-icon> Move to Top</v-list-item-title></v-list-item>
          <v-list-item v-if="!disableup" dense @click="move('up')"><v-list-item-title><v-icon>mdi-arrow-up</v-icon> Move Up</v-list-item-title></v-list-item>
          <v-list-item v-if="!disabledown" dense @click="move('down')"><v-list-item-title><v-icon>mdi-arrow-down</v-icon> Move Down</v-list-item-title></v-list-item>
          <v-list-item v-if="!disabledown" dense @click="move('bottom')"><v-list-item-title><v-icon>mdi-arrow-collapse-down</v-icon> Move to Bottom</v-list-item-title></v-list-item>
        </v-list>
      </v-menu>
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
    move (to: string) {
      this.$http.put(`queue/${this.index}`, new URLSearchParams({
        move: to
      }))
    }
  },

  props: ['disabledown', 'disableup', 'downloading', 'error', 'index', 'progress', 'title']
})
</script>
