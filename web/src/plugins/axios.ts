import axios from 'axios'
import Vue from 'vue'
import VueAxios from 'vue-axios'

const config = {
  baseURL: process.env.NODE_ENV === 'production' ? process.env.API_URL || '' : 'http://localhost'
}

const _axios = axios.create(config)

Vue.use(VueAxios, _axios)
