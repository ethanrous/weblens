import API_ENDPOINT from './ApiEndpoint.js'
import { ApiKeysApiFactory } from './swag/api.js'

const AccessApi = ApiKeysApiFactory(null, API_ENDPOINT)
export default AccessApi
