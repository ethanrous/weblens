import { WsMsg, WsMsgEvent } from '../../api/Websocket'
import { useSessionStore } from '../../components/UserInfo'
import { UserInfoT } from '../../types/Types'

export function setupWebsocketHandler(
    setRestoreInProgress: (prog: boolean) => void,
    nav: (loc: string) => void
) {
    return (msgData: WsMsg) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.WeblensLoadedEvent: {
                if (msgData.content['role'] === 'core') {
                    useSessionStore.getState().setUserInfo({} as UserInfoT)
                    useSessionStore
                        .getState()
                        .fetchServerInfo()
                        .then(() => {
                            nav('/files/home')
                        })
                } else if (msgData.content['role'] === 'restore') {
                    setRestoreInProgress(true)
                }
                break
            }
            case 'restore_started': {
                setRestoreInProgress(true)
                break
            }
            case 'going_down': {
                break
            }
            default: {
                console.error('Unknown websocket message', msgData.eventTag)
                break
            }
        }
    }
}
