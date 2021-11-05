<template>
  <v-list-item class="px-0 py-0" dense>
    <v-list-item-action class="mr-2 my-0 py-0">
      <v-btn depressed small class="text-none text-h6" color="deep-orange accent-1" @click="queue">Add</v-btn>
    </v-list-item-action>
    <v-list-item-action v-if="value.Video"><v-icon>mdi-video</v-icon></v-list-item-action>
    <v-list-item-content class="ml-1 my-0 py-0">
      <v-list-item-title class="my-0 py-0"><span>{{ value.Title }}</span></v-list-item-title>
    </v-list-item-content>
    <v-list-item-action class="my-0 py-0">
      <v-dialog v-model="dialog" :persistent="updating" max-width="900">
        <template v-slot:activator="{ on, attrs }">
          <v-btn plain small text :ripple="false" v-bind="attrs" v-on="on"><v-icon>mdi-pencil</v-icon></v-btn>
        </template>
        <v-card>
          <!-- we have to sacrifice vertical centering in exchange for actually having text-truncate work properly -->
          <!-- have i mentioned how much i hate webdev? -->
          <v-card-title><span class="text-h5 text-truncate"><v-icon large>mdi-{{ type }}</v-icon> {{ value.Title }}</span></v-card-title>
          <v-card-text>
            <v-text-field label="ID" dense :disabled="updating" outlined :readonly="!updating" :value="id"/>
            <v-text-field label="URL" dense :disabled="updating" outlined :readonly="!updating" :value="value.URL" append-icon="mdi-open-in-new" @click:append="openURL"/>
            <v-text-field label="Length" dense :disabled="updating" outlined :readonly="!updating" :value="$parseTime(this.value.Length / 1000000)"/>
            <v-text-field label="Created At" dense :disabled="updating" outlined :readonly="!updating" :value="new Date(value.CreatedAt).toLocaleString()"/>
            <v-text-field label="Last Updated" dense :disabled="updating" hide-details outlined :readonly="!updating" :value="new Date(value.UpdatedAt).toLocaleString()"/>
            <!-- okay... how in the bloody hell is ml-0 being explicitly there different from it not being there??? -->
            <v-row class="ml-0 mt-1">
              <v-switch v-model="video" class="mr-12" hide-details label="Video" :disabled="updating"/>
              <v-checkbox v-model="title" label="Update Title" class="mr-12" :disabled="updating" hide-details/>
              <v-checkbox v-model="length" label="Update Length" :disabled="updating" hide-details/>
            </v-row>
          </v-card-text>
          <v-card-actions>
            <v-spacer/>
            <v-btn class="text-none text-subtitle-1" color="deep-orange accent-1" depressed :disabled="updating" :loading="updating" @click="update">Update</v-btn>
            <v-btn class="text-none text-subtitle-1" color="secondary black--text" depressed :disabled="updating" @click="dialog = false">Close</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-list-item-action>
    <v-snackbar timeout="3000" v-model="addMessage">Added "{{ value.Title }}" to queue!</v-snackbar>
    <v-snackbar timeout="3000" v-model="updateMessage">{{ updateText }}</v-snackbar>
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
    addMessage: false,
    dialog: false,
    length: false,
    title: false,
    updateMessage: false,
    updateText: '',
    updating: false,
    video: false
  }),
  methods: {
    openURL () {
      window.open(this.value.URL, '_blank')
    },
    async queue (): Promise<void> {
      await this.$http.post('queue', new URLSearchParams({
        id: this.id,
        type: this.type
      }))
      this.addMessage = true
    },
    reset () {
      this.video = this.value.Video
      this.length = false
      this.title = false
    },
    async update () {
      this.updating = true
      try {
        const response = await this.$http.put(`media/${this.type}/${this.id}`, new URLSearchParams({
          length: this.length.toString(),
          title: this.title.toString(),
          video: this.video.toString()
        }))
        this.updateText = `Updated media metadata for "${this.id}"!`
        if (this.video !== this.value.Video && this.video && !response.data.Video) {
          this.updateText += ' Video is not supported and was ignored.'
        }
        this.$emit('input', response.data)
      } catch (e) {
        switch (e.response.status) {
          case 304:
            this.updateText = `Media metadata is up to date for "${this.id}".`
            if (this.video !== this.value.Video && this.video) {
              this.updateText += ' Video is not supported and was ignored.'
            }
            break
          case 404:
            this.updateText = 'Could not update media metadata: not found.'
            break
          case 500:
            this.updateText = 'Could not update media metadata: server error.'
            break
          default:
            this.updateText = 'Could not update media metadata: unknown error.'
            break
        }
      }
      this.reset()
      this.updateMessage = true
      this.updating = false
    }
  },
  mounted () {
    this.reset()
  },
  props: ['id', 'type', 'value'],
  watch: {
    dialog (newVal, oldVal) {
      if (!newVal && newVal !== oldVal) {
        this.reset()
      }
    }
  }
})
</script>
