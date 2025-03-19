import API_ENDPOINT from './ApiEndpoint.js'
import { ShareApiFactory } from './swag/api.js'

const SharesApi = ShareApiFactory(null, API_ENDPOINT)
export default SharesApi
