module.exports = {
  // yes, all of this is needed because vue-cli apparently overlooked all of this
  pages: {
    index: {
      entry: 'src/main.ts',
      title: '2E Music Memes'
    }
  },
  transpileDependencies: [
    'vuetify'
  ]
}
