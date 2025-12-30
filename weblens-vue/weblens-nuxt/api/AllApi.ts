import { type WLAPI, WeblensAPIFactory } from '@ethanrous/weblens-api'
const BASE_PATH = 'http://localhost:8080/api/v1'.replace(/\/+$/, '')

const WLAPI = shallowRef<WLAPI>()

export const API_ENDPOINT = ref<string>()

export function useWeblensAPI(): WLAPI {
    if (!WLAPI.value) {
        const pageOrigin = window?.location?.origin ?? 'http://localhost:3000'

        const basePath = BASE_PATH.slice(BASE_PATH.indexOf('api'))
        const apiEndpoint =
            process.env.NODE_ENV === 'development' && process.env.VITE_APP_API_ENDPOINT
                ? (process.env.VITE_APP_API_ENDPOINT as string)
                : `${pageOrigin}/${basePath}`

        API_ENDPOINT.value = apiEndpoint
        WLAPI.value = WeblensAPIFactory(apiEndpoint)
    }

    return WLAPI.value
}
