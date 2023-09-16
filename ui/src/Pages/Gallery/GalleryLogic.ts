import { fetchData } from '../../api/GalleryApi'

export function mediaReducer(state, action) {
    switch (action.type) {
        case 'add_media': {
            return {
                ...state,
                mediaList: action.mediaList,
                mediaIdMap: action.mediaIdMap,
                mediaCount: action.mediaList.length,
                hasMoreMedia: action.hasMoreMedia,
                dateMap: action.dateMap
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
                mediaList: [],
                mediaIdMap: {},
                datemap: {},
                mediaCount: 0,
                maxMediaCount: 100,
                includeRaw: !state.includeRaw
            }

        }
        case 'set_presentation': {
            document.documentElement.style.overflow = "hidden"

            return {
                ...state,
                presentingHash: action.presentingHash
            }
        }
        case 'set_presentation_ref': {
            return {
                ...state,
                presentingRef: action.ref
            }
        }
        case 'presentation_next': {
            let incBy = 0
            if (!state.mediaIdMap[state.mediaIdMap[state.presentingHash].next]?.next && state.hasMoreMedia && !(state.loading || state.maxMediaCount > state.mediaCount)) {
                incBy = 10
            }
            return {
                ...state,
                maxMediaCount: state.maxMediaCount + incBy,
                presentingHash: state.mediaIdMap[state.presentingHash].next ? state.mediaIdMap[state.presentingHash].next : state.presentingHash
            }
        }
        case 'presentation_previous': {
            return {
                ...state,
                presentingHash: state.mediaIdMap[state.presentingHash].previous ? state.mediaIdMap[state.presentingHash].previous : state.presentingHash
            }
        }
        case 'stop_presenting': {
            document.documentElement.style.overflow = "visible"
            state.presentingRef.current.scrollIntoView({ behavior: 'instant', block: 'center' })
            return {
                ...state,
                presentingHash: ""
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

export function handleScroll(e, dispatch) {
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