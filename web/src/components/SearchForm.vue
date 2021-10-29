<template>
  <v-card outlined>
    <v-card-title class="py-2">
      <span>Choose From Past Songs</span>
      <v-btn plain text :ripple="false" @click="show = !show">
        <v-icon v-if="show">mdi-chevron-up</v-icon>
        <v-icon v-else>mdi-chevron-down</v-icon>
      </v-btn>
    </v-card-title>
    <v-expand-transition>
      <div v-show="show">
        <v-card-text class="align-center d-flex my-0 py-0">
          <v-col class="mx-0 px-0 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 6 : 8">
            <v-text-field class="my-0 py-0" dense hide-details outlined label="Search" v-model="query" @keydown.enter.prevent="submit"/>
          </v-col>
          <v-col class="mx-0 px-2 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 3 : 2">
            <v-btn class="my-0 py-0 text-none text-h6" color="deep-orange accent-1" width="100%" depressed @click="submit">Search</v-btn>
          </v-col>
          <v-col class="mx-0 px-0 my-0 py-0" :cols="$vuetify.breakpoint.xs ? 3 : 2">
            <v-btn class="my-0 py-0 text-none text-h6" color="deep-orange accent-1" width="100%" depressed @click="getRandomSearch">Shuffle</v-btn>
          </v-col>
        </v-card-text>
        <v-card-text v-if="results.length">
          <SearchItem class="my-0 py-0" v-for="(item, i) in results" :key="i" :id="item.ID" :item="item" :title="item.Title" :type="item.Type"/>
        </v-card-text>
        <v-card-text v-else>
          <h3 class="text-center">No results found.</h3>
        </v-card-text>
      </div>
    </v-expand-transition>
  </v-card>
</template>

<style scoped>
/*
 * vuetify is missing a mechanism for setting card border colors
 */
.v-card {
  border-color: darkgray;
}
</style>

<script lang="ts">
import Vue from 'vue'
import SearchItem from './SearchItem.vue'

export default Vue.extend({
  name: 'SearchForm',

  components: {
    SearchItem
  },

  data: () => ({
    query: '',
    results: [],
    show: false
  }),

  methods: {
    async getRandomSearch () {
      this.results = (await this.$http.get('media/random', {
        params: {
          limit: 12
        }
      })).data
    },
    async submit (): Promise<void> {
      if (this.query.length) {
        this.results = (await this.$http.get('media/search', {
          params: {
            query: this.query
          }
        })).data
      } else {
        this.getRandomSearch()
      }
    }
  },

  mounted () {
    this.getRandomSearch()
  }
})
</script>
