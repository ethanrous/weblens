import { useEffect, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { EnqueueSnackbar, closeSnackbar } from 'notistack';
import { API_WS_ENDPOINT } from './ApiEndpoint'

export default function GetWebsocket(snacky: EnqueueSnackbar) {
    const [dcTimeout, setDcTimeout] = useState(null)
    const [dcSnack, setDcSnack] = useState(null)

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        onOpen: () => {
            clearTimeout(dcTimeout)
            if (dcSnack) {
                closeSnackbar(dcSnack)
                snacky("Websocket reconnected", { variant: "success" })
                setDcSnack(null)
            }
            console.log('WebSocket connection established.')
        },
        onClose: () => {
            if (!dcSnack && !dcTimeout) {
                setDcTimeout(setTimeout(() => {
                    setDcSnack(snacky("No connection to websocket, retrying...", { variant: "error", persist: true, preventDuplicate: true }))
                }, 2000))
            }
        },
        reconnectAttempts: 25,
        reconnectInterval: (last) => { return (last ^ 2) * 1000 },
        shouldReconnect: () => true,
        onReconnectStop: () => {
            clearTimeout(dcTimeout)
            closeSnackbar(dcSnack)
            setDcSnack(snacky("Unable to connect websocket. Please refresh your page", { variant: "error", persist: true, preventDuplicate: true }))
        }
    })
    let wsSend = (msg: string) => { sendMessage(msg) }
    return {
        wsSend, lastMessage, readyState
    }
}

export function dispatchSync(path, wsSend, recursive) {
    console.log("Doing sync")
    wsSend(JSON.stringify({
        type: 'scan_directory',
        content: {
            path: path,
            recursive: recursive
        },
    }))
}