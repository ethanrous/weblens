import API_ENDPOINT from './ApiEndpoint.js'
import { TowersApiFactory } from './swag/api.js'

export const TowersApi = TowersApiFactory(null, API_ENDPOINT)
