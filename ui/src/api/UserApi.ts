import API_ENDPOINT from './ApiEndpoint'
import { UsersApiFactory } from './swag'

const UsersApi = UsersApiFactory(null, API_ENDPOINT)
export default UsersApi
