import { useWebSocket } from '@vueuse/core'
import { defineStore } from 'pinia'
import { API_WS_ENDPOINT } from '~/api/ApiEndpoint'
import { handleWebsocketMessage } from '~/api/websocketHandlers'
import type { WsMessage } from '~/types/websocket'

export const useWebsocketStore = defineStore('websocket', () => {
    const {
        status,
        data,
        send: sendRaw,
        open,
        close,
    } = useWebSocket(API_WS_ENDPOINT, {
        autoReconnect: {
            retries: 3,
            delay: 1000,
            onFailed() {
                console.error('WebSocket connection failed after 3 retries')
            },
        },
    })

    watch(data, () => {
        const msg: WsMessage = JSON.parse(data.value)
        handleWebsocketMessage(msg)
    })

    function send(data: object) {
        const dataStr = JSON.stringify(data)
        sendRaw(dataStr)

        console.debug('Sent websocket message:', data)
    }

    return { status, data, send, open, close }
})
