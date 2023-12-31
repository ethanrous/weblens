import { useEffect } from 'react'
import { AlbumData, MediaData, MediaStateType, itemData } from '../../types/Types'
import { notifications } from '@mantine/notifications'

type galleryAction = {
    type: string
    media?: MediaData[]
    albums?: AlbumData[]
    block?: boolean
    itemId?: string
    item?: itemData
    progress?: number
    loading?: boolean
    search?: string
    open?: boolean
}

export function mediaReducer(state: MediaStateType, action: galleryAction) {
    switch (action.type) {
        case 'set_media': {
            state.mediaMap.clear()
            if (action.media) {
                let prev: MediaData
                for (const m of action.media) {
                    state.mediaMap.set(m.fileHash, m)
                    if (prev) {
                        prev.Next = m
                        m.Previous = prev
                    }
                    prev = m
                }
            }
            return {
                ...state
            }
        }

        case 'set_albums': {
            if (!action.albums) {
                return { ...state }
            }
            for (const a of action.albums) {
                state.albumsMap.set(a.Name, a)
            }
            return { ...state }
        }

        case 'set_block_search_focus': {
            return {
                ...state,
                blockSearchFocus: action.block
            }
        }

        case 'set_new_album_open': {
            return {
                ...state,
                blockSearchFocus: action.open,
                newAlbumDialogue: action.open
            }
        }

        case 'delete_from_map': {
            state.mediaMap.delete(action.itemId)
            // action.item
            return { ...state }
        }

        case 'set_scan_progress': {
            return {
                ...state,
                scanProgress: action.progress
            }
        }

        case 'set_loading': {
            return {
                ...state,
                loading: action.loading
            }
        }

        case 'toggle_raw': {
            window.scrollTo({
                top: 0,
                behavior: "smooth"
            })
            state.mediaMap.clear()
            return {
                ...state,
                mediaCount: 0,
                loading: true,
                includeRaw: !state.includeRaw
            }
        }

        case 'set_search': {
            return {
                ...state,
                searchContent: action.search,
            }
        }

        case 'set_presentation': {
            return {
                ...state,
                presentingHash: action.itemId
            }
        }

        case 'presentation_next': {
            return {
                ...state,
                presentingHash: state.mediaMap.get(state.presentingHash)?.Next ? state.mediaMap.get(state.presentingHash).Next.fileHash : state.presentingHash
            }
        }

        case 'presentation_previous': {
            return {
                ...state,
                presentingHash: state.mediaMap.get(state.presentingHash)?.Previous ? state.mediaMap.get(state.presentingHash).Previous.fileHash : state.presentingHash
            }
        }

        case 'stop_presenting': {
            if (state.presentingHash == "") {
                return {
                    ...state
                }
            }
            try {
                state.mediaMap.get(state.presentingHash).ImgRef.current.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'start' })
            } catch {
                console.error("No img ref: ", state.presentingHash)
            }
            return {
                ...state,
                presentingHash: ""
            }
        }

        default: {
            console.error("Do not have handler for dispatch type", action.type)
            return {
                ...state
            }
        }
    }
}

export const useKeyDown = (blockSearchFocus, searchRef) => {

    const onKeyDown = (event) => {
        if (!blockSearchFocus && !event.metaKey && ((event.which >= 65 && event.which <= 90) || event.key == "Backspace")) {
            searchRef.current.focus()
        } else if (event.key == "Escape") {
            searchRef.current.blur()
        }
    };
    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        };
    }, [onKeyDown])
}

export function handleWebsocket(lastMessage, dispatch) {
    if (lastMessage) {
        const msgData = JSON.parse(lastMessage.data)
        switch (msgData["type"]) {
            case "item_update": {
                return
            }
            case "item_deleted": {
                dispatch({ type: "delete_from_map", itemId: msgData["content"].hash })
                return
            }
            case "scan_directory_progress": {
                dispatch({ type: "set_scan_progress", progress: ((1 - (msgData["remainingTasks"] / msgData["totalTasks"])) * 100) })
                return
            }
            case "finished": {
                dispatch({ type: "set_loading", loading: false })
                return
            }
            case "error": {
                notifications.show({ message: msgData["error"], color: 'red' })
                return
            }
            default: {
                console.error("Got unexpected websocket message: ", msgData)
                return
            }
        }
    }
}
