import { useContext, useEffect, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { closeSnackbar, useSnackbar } from 'notistack';
import { API_WS_ENDPOINT } from './ApiEndpoint'
import { userContext } from '../Context';

export default function useWeblensSocket() {
    const [dcTimeout, setDcTimeout] = useState(null)
    const [dcSnack, setDcSnack] = useState(null)
    // const [wsData, setWsData] = useState({ wsSend: null, lastMessage: null, readyState: null })
    const { enqueueSnackbar } = useSnackbar()
    const { authHeader, userInfo } = useContext(userContext)
    // if (Object.keys(authHeader).length === 0) { return }

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        queryParams: authHeader,
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
        reconnectInterval: (last) => { return ((last + 1) ^ 2) * 1000 },
        shouldReconnect: () => true,
        onReconnectStop: () => {
            clearTimeout(dcTimeout)
            closeSnackbar(dcSnack)
            setDcSnack(enqueueSnackbar("Unable to connect to websocket. Please try refreshing your page", { variant: "error", persist: true, preventDuplicate: true }))
        }
    })
    let wsSend = (msg: string) => { sendMessage(msg) }
    // setWsData({ wsSend, lastMessage, readyState })

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