import API_ENDPOINT from './ApiEndpoint.js'
import { MediaApiFactory } from './swag/api.js'

const MediaApi = MediaApiFactory(null, API_ENDPOINT)
export default MediaApi
