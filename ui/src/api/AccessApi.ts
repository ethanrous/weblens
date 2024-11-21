import API_ENDPOINT from './ApiEndpoint'
import { ApiKeysApiFactory } from './swag'

const AccessApi = ApiKeysApiFactory(null, API_ENDPOINT)
export default AccessApi
