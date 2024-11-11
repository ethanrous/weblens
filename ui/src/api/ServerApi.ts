import API_ENDPOINT from './ApiEndpoint'
import { ServersApiFactory } from './swag'

const ServerApi = ServersApiFactory(null, API_ENDPOINT)
export default ServerApi
