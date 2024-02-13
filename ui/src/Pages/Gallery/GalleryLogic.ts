import { useCallback, useEffect } from 'react'
import { AlbumData, MediaData, MediaStateType, fileData } from '../../types/Types'
import { notifications } from '@mantine/notifications'

type galleryAction = {
    type: string
    medias?: MediaData[]
    albums?: AlbumData[]
    albumId?: string
    media?: MediaData
    albumNames?: string[]
    include?: boolean
    block?: boolean
    progress?: number
    loading?: boolean
    search?: string
    open?: boolean
    size?: number
    raw?: boolean
}

export function mediaReducer(state: MediaStateType, action: galleryAction): MediaStateType {
    switch (action.type) {
        case 'set_media': {
            state.mediaMap.clear()
            if (action.medias) {
                let prev: MediaData
                for (const m of action.medias) {
                    state.mediaMap.set(m.fileHash, m)
                    if (prev) {
                        prev.Next = m
                        m.Previous = prev
                    }
                    prev = m
                }
            }
            return {
                ...state,
                mediaMapUpdated: Date.now(),
                loading: false
            }
        }

        case 'set_albums': {
            if (!action.albums) {
                return { ...state }
            }
            state.albumsMap.clear()
            for (const a of action.albums) {
                state.albumsMap.set(a.Id, a)
            }
            return { ...state }
        }

        case 'set_album_media': {
            const album = state.albumsMap.get(action.albumId)
            album.CoverMedia = action.media
            state.albumsMap.set(action.albumId, album)
            return {...state}
        }

        case 'set_albums_filter': {
            return {
                ...state,
                albumsFilter: action.albumNames
            }
        }

        case 'set_image_size': {
            return {
                ...state,
                imageSize: action.size
            }
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
            state.mediaMap.delete(action.media.fileHash)
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

        case 'set_raw_toggle': {
            if (action.raw === state.includeRaw) {
                return {...state}
            }
            window.scrollTo({
                top: 0,
                behavior: "smooth"
            })
            state.mediaMap.clear()
            return {
                ...state,
                loading: true,
                includeRaw: action.raw
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
                presentingMedia: action.media
            }
        }

        case 'presentation_next': {
            return {
                ...state,
                presentingMedia: state.presentingMedia.Next ? state.presentingMedia.Next : state.presentingMedia
            }
        }

        case 'presentation_previous': {
            return {
                ...state,
                presentingMedia: state.presentingMedia.Previous ? state.presentingMedia.Previous : state.presentingMedia
            }
        }

        case 'stop_presenting': {
            if (state.presentingMedia === null) {
                return {
                    ...state
                }
            }
            try {
                state.presentingMedia.ImgRef.current.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'start' })
            } catch {
                console.error("No img ref: ", state.presentingMedia)
            }
            return {
                ...state,
                presentingMedia: null
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

    const onKeyDown = useCallback((event) => {
        if (!blockSearchFocus && !event.metaKey && ((event.which >= 65 && event.which <= 90) || event.key === "Backspace")) {
            searchRef.current.focus()
        } else if (event.key === "Escape") {
            searchRef.current.blur()
        }
    }, [blockSearchFocus, searchRef])
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
                dispatch({ type: "delete_from_map", media: msgData["content"].hash })
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
