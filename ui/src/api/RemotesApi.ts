import API_ENDPOINT from './ApiEndpoint'
import { RemotesApiFactory } from './swag'

export const RemoteApi = RemotesApiFactory(null, API_ENDPOINT)
