import { useCallback, useContext, useState } from 'react'
import useWebSocket from 'react-use-websocket'
import { API_WS_ENDPOINT } from './ApiEndpoint'
import { userContext } from '../Context'
import { notifications } from '@mantine/notifications'

export default function useWeblensSocket() {
    const [dcTimeout, setDcTimeout] = useState(null)
    const { userInfo, authHeader } = useContext(userContext)

    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        queryParams: authHeader,
        onOpen: () => {
            clearTimeout(dcTimeout)
            notifications.clean()
        },
        onClose: (event) => {
            if (!event.wasClean && authHeader && !dcTimeout && userInfo.username !== "") {
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
        shouldReconnect: () => userInfo.username !== "",
        onReconnectStop: () => {
            clearTimeout(dcTimeout)
            notifications.show({ id: "wsdc", message: "Websocket connection lost, please refresh your page", autoClose: false, color: 'red' })
        }
    })
    const wsSend = useCallback((action: string, content: any) => {
        const msg = {
            action: action,
            content: JSON.stringify(content)
        }
        console.log("WSSend", msg)
        sendMessage(JSON.stringify(msg))
    }, [sendMessage])

    return {
        wsSend, lastMessage, readyState
    }
}

export function dispatchSync(folderId: string, wsSend: (action: string, content: any) => void, recursive: boolean, full: boolean) {
    wsSend("scan_directory", {
            folderId: folderId,
            recursive: recursive,
            full: full
        }
    )
}