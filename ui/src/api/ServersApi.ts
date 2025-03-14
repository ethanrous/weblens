import API_ENDPOINT from './ApiEndpoint.js'
import { ServersApiFactory } from './swag/api.js'

export const ServersApi = ServersApiFactory(null, API_ENDPOINT)
