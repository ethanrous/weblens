import { createContext, useCallback, useEffect } from 'react'
import WeblensMedia from '../../Media/Media'
import {
    AlbumData,
    GalleryDispatchT,
    GalleryStateT,
    PresentType,
    TimeOffset,
} from '../../types/Types'

export type GalleryAction = {
    type: string
    medias?: WeblensMedia[]
    albums?: AlbumData[]
    albumId?: string
    mediaId?: string
    mediaIds?: string[]
    media?: WeblensMedia
    presentMode?: PresentType
    albumNames?: string[]
    include?: boolean
    block?: boolean
    progress?: number
    loading?: string
    search?: string
    selected?: boolean
    selecting?: boolean
    open?: boolean
    size?: number
    raw?: boolean
    targetId?: string
    pos?: { x: number; y: number }
    mediaIndex?: number
    shift?: boolean
    offset?: TimeOffset
}

export function galleryReducer(
    state: GalleryStateT,
    action: GalleryAction
): GalleryStateT {
    // console.log("Doing action!", action);

    switch (action.type) {
        // case 'set_media': {
        //     state.mediaMap.clear()
        //     if (action.medias) {
        //         let prev: WeblensMedia
        //         for (const m of action.medias) {
        //             state.mediaMap.set(m.Id(), m)
        //             if (prev) {
        //                 prev.SetNextLink(m)
        //                 m.SetPrevLink(prev)
        //             }
        //             prev = m
        //         }
        //     }
        //     return {
        //         ...state,
        //         mediaMap: new Map(state.mediaMap),
        //     }
        // }

        // case 'set_selected': {
        //     const item = state.mediaMap.get(action.mediaId)
        //     if (!item) {
        //         console.warn('Trying to select media that does not exist')
        //         return { ...state }
        //     }
        //     item.SetSelected(action.selected)
        //     if (state.holdingShift) {
        //         Array.from(state.mediaMap.values())
        //             .slice(
        //                 Math.min(action.mediaIndex, state.lastSelIndex),
        //                 Math.max(action.mediaIndex, state.lastSelIndex)
        //             )
        //             .map((v, i) => {
        //                 v.SetSelected(true)
        //                 state.selected.set(v.Id(), true)
        //             })
        //     }
        //
        //     let lastSel = state.lastSelIndex
        //     if (action.selected) {
        //         lastSel = action.mediaIndex
        //     }
        //
        //     return {
        //         ...state,
        //         selected: new Map(state.selected),
        //         lastSelIndex: lastSel,
        //     }
        // }

        // case 'set_selecting': {
        //     if (!action.selecting) {
        //         for (const key of state.mediaMap.keys()) {
        //             state.mediaMap.get(key).SetSelected(false)
        //         }
        //     }
        //     return {
        //         ...state,
        //         selecting: action.selecting,
        //     }
        // }

        case 'set_albums': {
            if (!action.albums) {
                return state
            }

            const newMap = new Map<string, AlbumData>()
            for (const a of action.albums) {
                newMap.set(a.id, a)
            }
            return { ...state, albumsMap: newMap }
        }

        case 'remove_album': {
            state.albumsMap.delete(action.albumId)
            return {
                ...state,
                albumsMap: new Map(state.albumsMap),
            }
        }

        case 'set_album_media': {
            const album = state.albumsMap.get(action.albumId)
            // album.cover = action.media
            state.albumsMap.set(action.albumId, album)
            return { ...state }
        }

        case 'set_albums_filter': {
            const albums = action.albumNames.filter((a) => {
                return Boolean(state.albumsMap.get(a))
            })

            if (state.albumsFilter.length === albums.length) {
                for (const a of albums) {
                    if (!state.albumsFilter.includes(a)) {
                        return {
                            ...state,
                            albumsFilter: albums,
                        }
                    }
                }
            } else {
                return {
                    ...state,
                    albumsFilter: albums,
                }
            }
            return state
        }

        case 'set_image_size': {
            if (!action.size) {
                return state
            }
            return {
                ...state,
                imageSize: action.size,
            }
        }

        case 'set_block_focus': {
            return {
                ...state,
                blockSearchFocus: action.block,
            }
        }

        // case 'delete_from_map': {
        //     for (const mId of action.mediaIds) {
        //         state.mediaMap.delete(mId)
        //         state.selected.delete(mId)
        //     }
        //
        //     return {
        //         ...state,
        //         mediaMap: new Map(state.mediaMap),
        //         selected: new Map(state.selected),
        //     }
        // }

        case 'add_loading': {
            const newLoading = state.loading.filter((v) => v !== action.loading)
            newLoading.push(action.loading)
            return {
                ...state,
                loading: newLoading,
            }
        }

        case 'remove_loading': {
            const newLoading = state.loading.filter((v) => v !== action.loading)
            return {
                ...state,
                loading: newLoading,
            }
        }

        case 'set_menu_target': {
            return { ...state, menuTargetId: action.targetId }
        }

        case 'set_raw_toggle': {
            if (action.raw === state.includeRaw) {
                return { ...state }
            }
            window.scrollTo({
                top: 0,
                behavior: 'smooth',
            })
            // state.mediaMap.clear()
            return {
                ...state,
                includeRaw: action.raw,
                // mediaMap: new Map(state.mediaMap),
            }
        }

        case 'set_search': {
            return {
                ...state,
                searchContent: action.search,
            }
        }

        case 'set_presentation': {
            if (!action.mediaId) {
                return state
            }
            if (action.presentMode && action.presentMode !== PresentType.None) {
                state.presentingMode = action.presentMode
            }
            return {
                ...state,
                presentingMediaId: action.mediaId,
            }
        }

        // case 'presentation_next': {
        //     let nextM = state.presentingMediaId.Next()
        //     if (state.presentingMode === PresentType.InLine && nextM) {
        //         nextM.GetImgRef().current.scrollIntoView({
        //             behavior: 'smooth',
        //             block: 'start',
        //             inline: 'start',
        //         })
        //     }
        //
        //     return {
        //         ...state,
        //         presentingMediaId: nextM ? nextM : state.presentingMediaId,
        //     }
        // }
        //
        // case 'presentation_previous': {
        //     return {
        //         ...state,
        //         presentingMediaId: state.presentingMediaId.Prev()
        //             ? state.presentingMediaId.Prev()
        //             : state.presentingMediaId,
        //     }
        // }

        case 'stop_presenting': {
            if (state.presentingMediaId === null) {
                return {
                    ...state,
                    presentingMode: PresentType.None,
                }
            }
            // try {
            //     state.presentingMediaId.GetImgRef().current.scrollIntoView({
            //         behavior: 'smooth',
            //         block: 'nearest',
            //         inline: 'start',
            //     })
            // } catch {
            //     console.error('No img ref: ', state.presentingMediaId)
            // }

            return {
                ...state,
                presentingMediaId: null,
                presentingMode: PresentType.None,
            }
        }

        case 'set_hover_target': {
            return { ...state, hoverIndex: action.mediaIndex }
        }

        case 'set_holding_shift': {
            return { ...state, holdingShift: action.shift }
        }

        case 'set_time_offset': {
            if (action.offset === null) {
                return { ...state, timeAdjustOffset: null }
            }
            return { ...state, timeAdjustOffset: { ...action.offset } }
        }

        default: {
            console.error('Do not have handler for dispatch type', action.type)
            return {
                ...state,
            }
        }
    }
}

export const useKeyDownGallery = (
    galleryState: GalleryStateT,
    galleryDispatch: GalleryDispatchT
) => {
    const onKeyDown = useCallback(
        (event) => {
            // if (event.key === 'Escape') {
            //     if (galleryState.menuTargetId) {
            //         return
            //     }
            //     galleryDispatch({ type: 'set_selecting', selecting: false })
            //     galleryDispatch({ type: 'stop_presenting' })
            if (event.key === 'Shift') {
                galleryDispatch({ type: 'set_holding_shift', shift: true })
            }
        },
        [
            galleryState?.blockSearchFocus,
            galleryState?.menuTargetId,
            galleryDispatch,
        ]
    )

    const onKeyUp = useCallback(
        (event) => {
            if (event.key === 'Shift') {
                galleryDispatch({ type: 'set_holding_shift', shift: false })
            }
        },
        [galleryState?.blockSearchFocus, galleryDispatch]
    )

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        document.addEventListener('keyup', onKeyUp)

        return () => {
            document.removeEventListener('keydown', onKeyDown)
            document.removeEventListener('keyup', onKeyUp)
        }
    }, [onKeyDown, onKeyUp])
}

export type GalleryContextT = {
    galleryState: GalleryStateT
    galleryDispatch: GalleryDispatchT
}

export const GalleryContext = createContext<GalleryContextT>({
    galleryState: null,
    galleryDispatch: null,
})
