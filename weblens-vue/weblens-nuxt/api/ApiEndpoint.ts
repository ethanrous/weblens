const BASE_PATH = 'http://localhost:8080/api/v1'.replace(/\/+$/, '')
const basePath = BASE_PATH.slice(BASE_PATH.indexOf('api'))

const API_ENDPOINT = ''
// process.env.NODE_ENV === 'development' && process.env.VITE_APP_API_ENDPOINT
//     ? (process.env.VITE_APP_API_ENDPOINT as string)
//     : `${window.location.origin}/${basePath}`

export const API_WS_ENDPOINT =
    process.env.NODE_ENV === 'development' && process.env.VITE_APP_API_WS_ENDPOINT
        ? (process.env.VITE_APP_API_WS_ENDPOINT as string)
        : window.location.protocol === 'https:'
          ? `wss://${window.location.host}/${basePath}/ws`
          : `ws://${window.location.host}/${basePath}/ws`

export default API_ENDPOINT
