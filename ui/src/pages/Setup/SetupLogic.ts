import { useSessionStore } from '@weblens/components/UserInfo'
import { UserInfoT } from '@weblens/types/Types'

export function setupWebsocketHandler(setRestoreInProgress, nav) {
    return (msgData) => {
        switch (msgData.eventTag) {
            case 'weblens_loaded': {
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
