<template>
  <v-list-item class="px-0 py-0" dense>
    <v-list-item-action class="my-0 py-0">
      <v-btn depressed small class="text-none text-h6" color="deep-orange accent-1" @click="queue">Add</v-btn>
    </v-list-item-action>
    <v-list-item-content class="my-0 py-0">
      <v-list-item-title class="my-0 py-0"><span>{{ title }}</span></v-list-item-title>
    </v-list-item-content>
    <v-snackbar timeout="3000" v-model="addMessage">Added "{{ title }}"  to queue!</v-snackbar>
  </v-list-item>
</template>

<style scoped>
/*
 * text classes aren't great for v-list-item-title
 */
span {
  font-size: 1.4em;
}
</style>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'SearchItem',

  data: () => ({
    addMessage: false
  }),

  methods: {
    async queue (): Promise<void> {
      await this.$http.post('queue', new URLSearchParams({
        id: this.item.ID,
        type: this.item.Type
      }))
      this.addMessage = true
    }
  },

  props: ['item', 'title']
})
</script>
