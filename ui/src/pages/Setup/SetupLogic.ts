import User from '@weblens/types/user/User'
import { WsMsgEvent, wsMsgInfo } from '../../api/Websocket'
import { useSessionStore } from '../../components/UserInfo'

export function setupWebsocketHandler(
    setRestoreInProgress: (prog: boolean) => void,
    nav: (loc: string) => void
) {
    return (msgData: wsMsgInfo) => {
        switch (msgData.eventTag) {
            case WsMsgEvent.WeblensLoadedEvent: {
                if (msgData.content['role'] === 'core') {
                    useSessionStore.getState().setUser(new User({}, false))
                    useSessionStore
                        .getState()
                        .fetchServerInfo()
                        .then(() => {
                            nav('/files/home')
                        })
                        .catch((err) => {
                            console.error('Failed to fetch server info', err)
                        })
                } else if (msgData.content['role'] === 'restore') {
                    setRestoreInProgress(true)
                }
                break
            }
            case WsMsgEvent.RestoreStartedEvent: {
                setRestoreInProgress(true)
                break
            }
            case WsMsgEvent.ServerGoingDownEvent: {
                break
            }
            default: {
                console.error('Unknown websocket message', msgData.eventTag)
                break
            }
        }
    }
}
