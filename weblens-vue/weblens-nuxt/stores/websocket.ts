import { useWebSocket } from '@vueuse/core'
import { defineStore } from 'pinia'
import { API_WS_ENDPOINT } from '~/api/ApiEndpoint'
import { handleWebsocketMessage } from '~/api/websocketHandlers'
import type { WsMessage } from '~/types/websocket'
import { useReconnectListener } from '~/composables/useReconnectListener'

const useWebsocketStore = defineStore('websocket', () => {
    const userStore = useUserStore()
    const openedOnce = ref(false)

    const {
        status,
        data,
        send: sendRaw,
        open,
        close,
    } = useWebSocket(API_WS_ENDPOINT, {
        immediate: userStore.user.isLoggedIn.get({ default: false }),
        autoReconnect: {
            retries: 3,
            delay: 1000,
            onFailed() {
                console.error('WebSocket connection failed after 3 retries')
            },
        },
    })

    const hasConnected = ref(false)

    watch(data, () => {
        const msg: WsMessage = JSON.parse(data.value)
        handleWebsocketMessage(msg)
    })

    watchEffect(() => {
        // Automatically open websocket when user logs in. This might not happen on initial load
        // if the user store loads after the websocket store (likely).
        if (userStore.user.isLoggedIn.get({ default: false }) && status.value !== 'OPEN' && !openedOnce.value) {
            openedOnce.value = true
            open()
        }
    })

    watch(status, (newStatus) => {
        if (newStatus !== 'OPEN') {
            return
        }

        if (hasConnected.value) {
            console.debug('WebSocket reconnected, replaying subscriptions')
            useReconnectListener().replay()
        }

        hasConnected.value = true
    })

    function send(data: object) {
        const dataStr = JSON.stringify(data)
        sendRaw(dataStr)

        console.debug('Sent websocket message:', data)
    }

    return { status, data, send, open, close }
})

export default useWebsocketStore
