import { reactive } from 'vue'
import { client } from './client.js'

export const initState = reactive({
  ready: false,
  version: '',
  siteTitle: 'Greyvar Dat Editor',
  datDir: '',
  resDir: '',
  error: '',
})

export async function loadInit () {
  try {
    const res = await client.init({})
    initState.version = res.version
    initState.siteTitle = res.siteTitle
    initState.datDir = res.datDir
    initState.resDir = res.resDir
    initState.error = ''
    document.title = res.siteTitle
  } catch (err) {
    initState.error = err.message || 'Failed to connect'
  } finally {
    initState.ready = true
  }
}
