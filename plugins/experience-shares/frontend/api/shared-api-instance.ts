import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 60000
})

export default api
