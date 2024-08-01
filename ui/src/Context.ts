import { createContext } from 'react'
import { MediaAction, MediaStateT } from './Media/Media'

export const MediaContext = createContext<{
    mediaState: MediaStateT
    mediaDispatch: (mediaAction: MediaAction) => void
}>(null)
export const WebsocketContext =
    createContext<(actionKey: string, content: any) => void>(null)
