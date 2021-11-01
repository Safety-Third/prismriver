<template>
  <v-list-item class="px-0 py-0" dense>
    <v-list-item-action class="mr-2 my-0 py-0">
      <v-btn depressed small class="text-none text-h6" color="deep-orange accent-1" @click="queue">Add</v-btn>
    </v-list-item-action>
    <v-list-item-action v-if="video"><v-icon>mdi-video</v-icon></v-list-item-action>
    <v-list-item-content class="ml-1 my-0 py-0">
      <v-list-item-title class="my-0 py-0"><span>{{ title }}</span></v-list-item-title>
    </v-list-item-content>
    <v-list-item-action class="my-0 py-0">
      <v-dialog v-model="dialog" :persistent="updatingVideo" max-width="900">
        <template v-slot:activator="{ on, attrs }">
          <v-btn plain small text :ripple="false" v-bind="attrs" v-on="on"><v-icon>mdi-pencil</v-icon></v-btn>
        </template>
        <v-card>
          <v-card-title><v-icon large class="mr-1">mdi-{{ type }}</v-icon><span class="text-h5 text-truncate">{{ title }}</span></v-card-title>
          <v-card-text>
            <v-text-field label="ID" dense outlined readonly :value="id"/>
            <v-text-field label="Length" dense outlined readonly :value="$parseTime(this.item.Length / 1000000)"/>
            <v-text-field label="Created At" dense outlined readonly :value="new Date(item.CreatedAt).toLocaleString()"/>
            <v-text-field label="Last Updated" dense hide-details outlined readonly :value="new Date(item.UpdatedAt).toLocaleString()"/>
            <v-switch v-model="video" hide-details label="Video" :disabled="updatingVideo" :loading="updatingVideo" @change="updateVideo"/>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn class="text-none text-subtitle-1" color="deep-orange accent-1" depressed :disabled="updatingVideo" @click="dialog = false">Close</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-list-item-action>
    <v-snackbar timeout="3000" v-model="addMessage">Added "{{ title }}" to queue!</v-snackbar>
    <v-snackbar timeout="3000" v-model="videoMessage">{{ videoText }}</v-snackbar>
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

  computed: {
    // this really actively goes out of its way to break everything, doesn't it?
    videoText (): string {
      switch (this.response) {
        case 204:
        case 304:
          return `Updated video setting for "${this.title}"!`
        case 400:
          return 'Could not update video setting: invalid input.'
        case 403:
          return 'Could not update video setting: video not supported.'
        case 404:
          return 'Could not update video setting: media not found.'
        case 500:
          return 'Could not update video setting: downloader is not working.'
        default:
          throw new TypeError(`unexpected response code ${this.response} from server`)
      }
    }
  },
  data: () => ({
    addMessage: false,
    dialog: false,
    response: 204,
    updatingVideo: false,
    video: false,
    videoMessage: false
  }),
  methods: {
    async queue (): Promise<void> {
      await this.$http.post('queue', new URLSearchParams({
        id: this.id,
        type: this.type
      }))
      this.addMessage = true
    },
    async updateVideo () {
      this.updatingVideo = true
      try {
        await this.$http.put(`media/${this.type}/${this.id}`, new URLSearchParams({
          video: this.video.toString()
        }))
        this.response = 204
      } catch (e) {
        this.response = e.response.status
        this.video = !this.video
      }
      this.videoMessage = true
      this.updatingVideo = false
    }
  },
  mounted () {
    this.video = this.item.Video
  },
  props: ['id', 'item', 'title', 'type']
})
</script>
