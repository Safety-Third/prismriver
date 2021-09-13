<template>
  <v-card outlined>
    <v-card-title class="py-2">Queue Music - Enter a link or go random!</v-card-title>
    <v-card-text class="align-center d-flex my-0 py-0">
      <v-col class="mx-0 px-0 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 6 : 8">
        <v-text-field class="my-0 py-0" dense hide-details outlined label="Music URL" v-model="url" @keydown.enter.prevent="submit"/>
      </v-col>
      <v-col class="mx-0 px-2 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 3 : 2">
        <v-btn depressed class="my-0 py-0 text-none text-h6" color="deep-orange accent-1" width="100%" @click="submit">Add</v-btn>
      </v-col>
      <v-col class="mx-0 px-0 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 3 : 2">
        <v-btn depressed class="my-0 py-0 text-none text-h6" color="deep-orange accent-1" width="100%" @click="random">Random</v-btn>
      </v-col>
    </v-card-text>
    <v-card-actions class="mb-0 pb-0 mt-0 pt-0">
      <v-checkbox class="mb-0 pb-0 mt-1 pt-1 ml-2" label="With Video" v-model="video"/>
    </v-card-actions>
    <v-snackbar timeout="3000" v-model="addMessage">Submitted! Now preparing song...</v-snackbar>
    <v-snackbar timeout="3000" v-model="randomMessage">Added a random song!</v-snackbar>
  </v-card>
</template>

<style scoped>
/*
 * v-card doesn't have a means of setting border color
 */
.v-card {
  border-color: darkgray;
}
</style>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'URLForm',

  data: () => ({
    addMessage: false,
    randomMessage: false,
    video: false,
    url: ''
  }),

  methods: {
    async random () {
      const data = (await this.axios.get('media/random', {
        params: {
          limit: 1
        }
      })).data[0]
      await this.$http.post('queue', new URLSearchParams({
        id: data.ID,
        type: data.Type
      }))
      this.randomMessage = true
    },
    async submit () {
      await this.$http.post('queue', new URLSearchParams({
        url: this.url,
        video: this.video.toString()
      }))
      this.addMessage = true
      this.url = ''
      this.video = false
    }
  }
})
</script>
