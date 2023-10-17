import excludeVariablesFromRoot from '@mui/material/styles/excludeVariablesFromRoot'
import { fetchData } from '../../api/GalleryApi'
import { MediaData } from '../../types/Generic'

export function mediaReducer(state, action) {
    switch (action.type) {
        case 'add_media': {
            return {
                ...state,
                mediaMap: action.mediaMap,
                dateMap: action.dateMap,
                hasMoreMedia: action.hasMoreMedia,
                previousLast: action.previousLast,
                mediaCount: state.mediaCount + action.addedCount
            }
        }
        case 'insert_thumbnail': {
            state.mediaMap[action.hash].Thumbnail64 = action.thumb64
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
        case 'insert_fullres': {
            state.mediaMap[action.hash].Fullres64 = action.fullres64
            return {
                ...state,
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
        case 'toggle_info': {
            return {
                ...state,
                showIcons: !state.showIcons
            }
        }
        case 'toggle_raw': {
            window.scrollTo({
                top: 0,
                behavior: "smooth"
            })
            return {
                ...state,
                mediaMap: {},
                dateMap: {},
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
            console.log("Setting: ", action.presentingHash)
            document.documentElement.style.overflow = "hidden"

            return {
                ...state,
                presentingHash: action.presentingHash
            }
        }
        case 'presentation_next': {
            console.log("Here1")
            let incBy = 0
            if (!state.mediaMap[state.presentingHash]?.next?.next && state.hasMoreMedia && !(state.loading || state.maxMediaCount > state.mediaCount)) {
                incBy = 100
            }
            return {
                ...state,
                maxMediaCount: state.maxMediaCount + incBy,
                presentingHash: state.mediaMap[state.presentingHash]?.next ? state.mediaMap[state.presentingHash].next.FileHash : state.presentingHash
            }
        }
        case 'presentation_previous': {
            return {
                ...state,
                presentingHash: state.mediaMap[state.presentingHash].previous ? state.mediaMap[state.presentingHash].previous.FileHash : state.presentingHash
            }
        }
        case 'stop_presenting': {
            if (state.presentingHash == "") {
                return {
                    ...state
                }
            }
            document.documentElement.style.overflow = "visible"
            try {
                state.mediaMap[state.presentingHash].ImgRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
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

export function moreData(mediaState, dispatch) {
    if (mediaState.loading) {
        return
    }

    dispatch({ type: "set_loading", loading: true })
    fetchData(mediaState, dispatch).then(() => dispatch({ type: "set_loading", loading: false }))

}