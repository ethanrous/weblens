import { useContext, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { closeSnackbar, useSnackbar } from 'notistack';
import { API_WS_ENDPOINT } from './ApiEndpoint'
import { userContext } from '../Context';
import { notifications } from '@mantine/notifications';

export default function useWeblensSocket() {
    const [dcTimeout, setDcTimeout] = useState(null)
    // const [wsData, setWsData] = useState({ wsSend: null, lastMessage: null, readyState: null })
    const { enqueueSnackbar } = useSnackbar()
    const { authHeader, userInfo } = useContext(userContext)
    // if (Object.keys(authHeader).length === 0) { return }

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        queryParams: authHeader,
        onOpen: () => {
            clearTimeout(dcTimeout)
            closeSnackbar()
            notifications.clean()
            // enqueueSnackbar("Websocket reconnected", { variant: "success" })

            console.log('WebSocket connection established.')
        },
        onClose: (event) => {
            if (!event.wasClean && authHeader && !dcTimeout) {
                setDcTimeout(setTimeout(() => {
                    notifications.show({
                        id: "wsdc",
                        message: "Lost websocket connection, retrying...",
                        color: "red",
                        // icon: <IconCheck style={{ width: rem(18), height: rem(18) }} />,
                        loading: false
                    })
                }, 2000))
            }
        },
        reconnectAttempts: 5,
        reconnectInterval: (last) => { return ((last + 1) ^ 2) * 1000 },
        shouldReconnect: () => true,
        onReconnectStop: () => {
            clearTimeout(dcTimeout)
            notifications.show({ id: "wsdc", message: "Lost websocket connection, retrying...", autoClose: false })
        }
    })
    let wsSend = (msg: string) => { sendMessage(msg) }
    // setWsData({ wsSend, lastMessage, readyState })

    return {
        wsSend, lastMessage, readyState
    }
}

export function dispatchSync(folderId: string, wsSend: (msg: string) => void, recursive: boolean) {
    console.log("SCANNING")
    wsSend(JSON.stringify({
        req: 'scan_directory',
        content: {
            folderId: folderId,
            recursive: recursive
        },
    }))
}