const API_ENDPOINT = (process.env.NODE_ENV === 'development' && process.env.REACT_APP_API_ENDPOINT)
    ? (process.env.REACT_APP_API_ENDPOINT as string)
    : `${window.location.origin}/api`

export const API_WS_ENDPOINT = (process.env.NODE_ENV === 'development' && process.env.REACT_APP_API_WS_ENDPOINT)
    ? (process.env.REACT_APP_API_WS_ENDPOINT as string)
    : window.location.protocol === "https:" ? `wss://${window.location.host}/api/ws` : `ws://${window.location.host}/api/ws`

export default API_ENDPOINT