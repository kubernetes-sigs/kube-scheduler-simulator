import axios from 'axios'

export const baseURL = process.env.BASE_URL + '/api/v1'
export const instance = axios.create({
  baseURL,
  withCredentials: true,
})
