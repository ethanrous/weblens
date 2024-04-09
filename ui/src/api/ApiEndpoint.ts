const API_ENDPOINT =
    import.meta.env.NODE_ENV === "development" && import.meta.env.VITE_APP_API_ENDPOINT
        ? (import.meta.env.VITE_APP_API_ENDPOINT as string)
        : `${window.location.origin}/api`;

export const ADMIN_ENDPOINT = `${API_ENDPOINT}/admin`;
export const PUBLIC_ENDPOINT = `${API_ENDPOINT}/public`;

export const API_WS_ENDPOINT =
    import.meta.env.NODE_ENV === "development" && import.meta.env.VITE_APP_API_WS_ENDPOINT
        ? (import.meta.env.VITE_APP_API_WS_ENDPOINT as string)
        : window.location.protocol === "https:"
          ? `wss://${window.location.host}/api/ws`
          : `ws://${window.location.host}/api/ws`;

export default API_ENDPOINT;
