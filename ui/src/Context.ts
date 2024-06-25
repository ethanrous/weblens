import { createContext } from 'react'
import { MediaAction, MediaStateT } from './Media/Media'
import { UserContextT } from './types/Types'

export const UserContext = createContext<UserContextT>(null)
export const MediaContext = createContext<{
    mediaState: MediaStateT
    mediaDispatch: (mediaAction: MediaAction) => void
}>(null)
export const WebsocketContext = createContext<(s: string, c: any) => void>(null)
