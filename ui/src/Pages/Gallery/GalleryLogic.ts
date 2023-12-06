import { useEffect } from 'react'
import { MediaStateType, itemData } from '../../types/Types'

export function mediaReducer(state: MediaStateType, action) {
    switch (action.type) {
        case 'add_media': {
            return {
                ...state,
                mediaMap: action.mediaMap,
                hasMoreMedia: action.hasMoreMedia,
                previousLast: action.previousLast,
                mediaCount: state.mediaCount + action.addedCount,
            }
        }

        case 'delete_from_map': {
            state.mediaMap.delete(action.item)
            // action.item
            return { ...state }
        }

        case 'insert_thumbnail': {
            state.mediaMap.get(action.hash).Thumbnail64 = action.thumb64
            return {
                ...state,
            }
        }

        case 'set_scan_progress': {
            return {
                ...state,
                scanProgress: action.progress
            }
        }

        case 'inc_max_media_count': {
            if (state.loading || state.maxMediaCount > state.mediaCount) {
                return {
                    ...state
                }
            }
            return {
                ...state,
                maxMediaCount: state.maxMediaCount + action.incBy,
                loading: true
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
                maxMediaCount: 100,
                hasMoreMedia: true,
                previousLast: "",
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
                presentingHash: action.presentingHash
            }
        }

        case 'presentation_next': {
            let incBy = 0
            if (!state.mediaMap.get(state.presentingHash)?.Next?.Next && state.hasMoreMedia && !(state.loading || state.maxMediaCount > state.mediaCount)) {
                incBy = 100
            }
            return {
                ...state,
                maxMediaCount: state.maxMediaCount + incBy,
                presentingHash: state.mediaMap.get(state.presentingHash)?.Next ? state.mediaMap.get(state.presentingHash).Next.FileHash : state.presentingHash
            }
        }

        case 'presentation_previous': {
            return {
                ...state,
                presentingHash: state.mediaMap.get(state.presentingHash)?.Previous ? state.mediaMap.get(state.presentingHash).Previous.FileHash : state.presentingHash
            }
        }

        case 'stop_presenting': {
            if (state.presentingHash == "") {
                return {
                    ...state
                }
            }
            try {
                state.mediaMap.get(state.presentingHash).ImgRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
            } catch {
                console.log("No img ref: ", state.presentingHash)
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

export const useKeyDown = (searchRef) => {

    const onKeyDown = (event) => {
        if (!event.metaKey && ((event.which >= 65 && event.which <= 90) || event.key == "Backspace")) {
            searchRef.current.children[0].focus()
        } else if (event.key == "Escape") {
            searchRef.current.children[0].blur()
        }
    };
    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        };
    }, [onKeyDown])
}

export const useScroll = (hasMoreMedia, dispatch) => {
    const onScrollEvent = (_) => {
        if (hasMoreMedia) { handleScroll(dispatch) }
    }
    useEffect(() => {
        window.addEventListener('scroll', onScrollEvent)
        return () => {
            window.removeEventListener('scroll', onScrollEvent)
        }
    }, [onScrollEvent])
}

export function handleScroll(dispatch) {
    console.log("HERE")
    if (document.documentElement.scrollHeight - (document.documentElement.scrollTop + window.innerHeight) < 1500) {

    }
}