import { MediaStateType } from '../../types/Types'

export function mediaReducer(state: MediaStateType, action) {
    switch (action.type) {
        case 'add_media': {
            return {
                ...state,
                mediaMap: action.mediaMap,
                hasMoreMedia: action.hasMoreMedia,
                previousLast: action.previousLast,
                mediaCount: state.mediaCount + action.addedCount
            }
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
                maxMediaCount: state.maxMediaCount + action.incBy
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
                previousLast: "",
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
            document.documentElement.style.overflow = "hidden"

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
                presentingHash: state.mediaMap.get(state.presentingHash).Previous ? state.mediaMap.get(state.presentingHash).Previous.FileHash : state.presentingHash
            }
        }

        case 'stop_presenting': {
            document.documentElement.style.overflow = "visible"
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

export function startKeybaordListener(dispatch) {

    const keyDownHandler = event => {
        if (event.key === 'i') {
            event.preventDefault()
            dispatch({
                type: 'toggle_info'
            })
        }
    }

    document.addEventListener('keydown', keyDownHandler)
    return () => {
        document.removeEventListener('keydown', keyDownHandler)
    }
}

export function handleScroll(dispatch) {
    if (document.documentElement.scrollHeight - (document.documentElement.scrollTop + window.innerHeight) < 1500) {
        dispatch({ type: "inc_max_media_count", incBy: 100 })
    }
}