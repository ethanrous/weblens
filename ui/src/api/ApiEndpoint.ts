import { BASE_PATH } from './swag/base';

const basePath = BASE_PATH.slice(BASE_PATH.indexOf('api'))

const API_ENDPOINT =
	import.meta.env.NODE_ENV === 'development' &&
		import.meta.env.VITE_APP_API_ENDPOINT
		? (import.meta.env.VITE_APP_API_ENDPOINT as string)
		: `${window.location.origin}/${basePath}`

export const API_WS_ENDPOINT =
	import.meta.env.NODE_ENV === 'development' &&
		import.meta.env.VITE_APP_API_WS_ENDPOINT
		? (import.meta.env.VITE_APP_API_WS_ENDPOINT as string)
		: window.location.protocol === 'https:'
			? `wss://${window.location.host}/${basePath}/ws`
			: `ws://${window.location.host}/${basePath}/ws`

export default API_ENDPOINT
