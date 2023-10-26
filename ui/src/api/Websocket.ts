import { useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { closeSnackbar, useSnackbar } from 'notistack';
import { API_WS_ENDPOINT } from './ApiEndpoint'

export default function useWeblensSocket() {
    const [dcTimeout, setDcTimeout] = useState(null)
    const [dcSnack, setDcSnack] = useState(null)
    const { enqueueSnackbar } = useSnackbar()

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        onOpen: () => {
            clearTimeout(dcTimeout)
            closeSnackbar()
            // enqueueSnackbar("Websocket reconnected", { variant: "success" })
            setDcSnack(null)

            console.log('WebSocket connection established.')
        },
        onClose: (event) => {

            if (!event.wasClean && !dcSnack && !dcTimeout) {
                setDcTimeout(setTimeout(() => {
                    setDcSnack(enqueueSnackbar("No connection to websocket, retrying...", { variant: "error", preventDuplicate: true }))
                }, 2000))
            }
        },
        reconnectAttempts: 5,
        reconnectInterval: (last) => { return (last ^ 2) * 1000 },
        shouldReconnect: () => true,
        onReconnectStop: () => {
            clearTimeout(dcTimeout)
            closeSnackbar(dcSnack)
            setDcSnack(enqueueSnackbar("Unable to connect websocket. Please refresh your page", { variant: "error", persist: true, preventDuplicate: true }))
        }
    })
    let wsSend = (msg: string) => { sendMessage(msg) }
    return {
        wsSend, lastMessage, readyState
    }
}

export function dispatchSync(path: string, wsSend: (msg: string) => void, recursive: boolean) {
    wsSend(JSON.stringify({
        type: 'scan_directory',
        content: {
            path: path,
            recursive: recursive
        },
    }))
}