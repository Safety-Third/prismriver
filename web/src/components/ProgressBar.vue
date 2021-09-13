<template>
  <v-progress-linear class="mt-4" height="20" :color="color" :value="value">{{ message }}</v-progress-linear>
</template>

<script lang="ts">
import Vue from 'vue'

export default Vue.extend({
  name: 'ProgressBar',

  computed: {
    color () {
      if (this.error) {
        return 'error'
      } else if (this.progress < 50) {
        return 'info'
      } else {
        return 'success'
      }
    },
    message () {
      if (this.error) {
        return this.error
      } else if (this.progress < 50) {
        return `Downloading (${Math.floor(this.progress)}%)`
      } else {
        return `Transcoding (${Math.floor(this.progress)}%)`
      }
    },
    value () {
      if (this.error) {
        return 100
      }
      return this.progress
    }
  },

  props: ['error', 'progress']
})
</script>
