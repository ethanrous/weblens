import { useContext, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { API_WS_ENDPOINT } from './ApiEndpoint'
import { userContext } from '../Context'
import { notifications } from '@mantine/notifications'

export default function useWeblensSocket() {
    const [dcTimeout, setDcTimeout] = useState(null)
    const { authHeader } = useContext(userContext)

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        queryParams: authHeader,
        onOpen: () => {
            clearTimeout(dcTimeout)
            notifications.clean()
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
    return {
        wsSend, lastMessage, readyState
    }
}

export function dispatchSync(folderId: string, wsSend: (msg: string) => void, recursive: boolean) {
    wsSend(JSON.stringify({
        req: 'scan_directory',
        content: {
            folderId: folderId,
            recursive: recursive
        },
    }))
}