import API_ENDPOINT from './ApiEndpoint.js'
import { UsersApiFactory } from './swag/api.js'

const UsersApi = UsersApiFactory(null, API_ENDPOINT)
export default UsersApi
