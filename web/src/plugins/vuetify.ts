import Vue from 'vue'
import { colors } from 'vuetify/lib'
import Vuetify from 'vuetify/lib/framework'

Vue.use(Vuetify)

export default new Vuetify({
  theme: {
    themes: {
      light: {
        primary: colors.deepOrange,
        secondary: colors.deepOrange.lighten4,
        accent: colors.deepOrange.accent1
      }
    }
  }
})
