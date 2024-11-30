import { AlbumInfo } from '@weblens/api/swag'
import { PresentType, TimeOffset } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { clamp } from '@weblens/util'
import { useCallback, useEffect } from 'react'
import { StateCreator, create } from 'zustand'
import { devtools } from 'zustand/middleware'

// export type GalleryAction = {
//     type: string
//     medias?: WeblensMedia[]
//     albums?: AlbumInfo[]
//     albumId?: string
//     mediaId?: string
//     mediaIds?: string[]
//     media?: WeblensMedia
//     presentMode?: PresentType
//     albumNames?: string[]
//     include?: boolean
//     block?: boolean
//     progress?: number
//     loading?: string
//     search?: string
//     selected?: boolean
//     selecting?: boolean
//     open?: boolean
//     size?: number
//     targetId?: string
//     pos?: { x: number; y: number }
//     mediaIndex?: number
//     shift?: boolean
//     offset?: TimeOffset
// }
//
// export function galleryReducer(
//     state: GalleryStateT,
//     action: GalleryAction
// ): GalleryStateT {
//     switch (action.type) {
//         case 'set_selecting': {
//             return {
//                 ...state,
//                 selecting: action.selecting,
//             }
//         }
//
//         case 'set_albums': {
//             if (!action.albums) {
//                 return state
//             }
//
//             const newMap = new Map<string, AlbumInfo>()
//             for (const a of action.albums) {
//                 newMap.set(a.id, a)
//             }
//             return { ...state, albumsMap: newMap }
//         }
//
//         case 'remove_album': {
//             state.albumsMap.delete(action.albumId)
//             return {
//                 ...state,
//                 albumsMap: new Map(state.albumsMap),
//             }
//         }
//
//         case 'set_album_media': {
//             const album = state.albumsMap.get(action.albumId)
//             // album.cover = action.media
//             state.albumsMap.set(action.albumId, album)
//             return { ...state }
//         }
//
//         case 'set_albums_filter': {
//             const albums = action.albumNames.filter((a) => {
//                 return Boolean(state.albumsMap.get(a))
//             })
//
//             if (state.albumsFilter.length === albums.length) {
//                 for (const a of albums) {
//                     if (!state.albumsFilter.includes(a)) {
//                         return {
//                             ...state,
//                             albumsFilter: albums,
//                         }
//                     }
//                 }
//             } else {
//                 return {
//                     ...state,
//                     albumsFilter: albums,
//                 }
//             }
//             return state
//         }
//
//         case 'set_image_size': {
//             if (!action.size) {
//                 return state
//             }
//             return {
//                 ...state,
//                 imageSize: action.size,
//             }
//         }
//
//         case 'set_block_focus': {
//             return {
//                 ...state,
//                 blockSearchFocus: action.block,
//             }
//         }
//
//         // case 'delete_from_map': {
//         //     for (const mId of action.mediaIds) {
//         //         state.mediaMap.delete(mId)
//         //         state.selected.delete(mId)
//         //     }
//         //
//         //     return {
//         //         ...state,
//         //         mediaMap: new Map(state.mediaMap),
//         //         selected: new Map(state.selected),
//         //     }
//         // }
//
//         case 'add_loading': {
//             const newLoading = state.loading.filter((v) => v !== action.loading)
//             newLoading.push(action.loading)
//             return {
//                 ...state,
//                 loading: newLoading,
//             }
//         }
//
//         case 'remove_loading': {
//             const newLoading = state.loading.filter((v) => v !== action.loading)
//             return {
//                 ...state,
//                 loading: newLoading,
//             }
//         }
//
//         case 'set_menu_target': {
//             return { ...state, menuTargetId: action.targetId }
//         }
//
//         case 'set_search': {
//             return {
//                 ...state,
//                 searchContent: action.search,
//             }
//         }
//
//         case 'set_presentation': {
//             if (!action.mediaId) {
//                 return { ...state, presentingMode: PresentType.None }
//             }
//             if (action.presentMode && action.presentMode !== PresentType.None) {
//                 state.presentingMode = action.presentMode
//             }
//             return {
//                 ...state,
//                 presentingMediaId: action.mediaId,
//             }
//         }
//
//         case 'stop_presenting': {
//             if (state.presentingMediaId === null) {
//                 return {
//                     ...state,
//                     presentingMode: PresentType.None,
//                 }
//             }
//             // try {
//             //     state.presentingMediaId.GetImgRef().current.scrollIntoView({
//             //         behavior: 'smooth',
//             //         block: 'nearest',
//             //         inline: 'start',
//             //     })
//             // } catch {
//             //     console.error('No img ref: ', state.presentingMediaId)
//             // }
//
//             return {
//                 ...state,
//                 presentingMediaId: null,
//                 presentingMode: PresentType.None,
//             }
//         }
//
//         case 'set_hover_target': {
//             return { ...state, hoverIndex: action.mediaIndex }
//         }
//
//         case 'set_holding_shift': {
//             return { ...state, holdingShift: action.shift }
//         }
//
//         case 'set_time_offset': {
//             if (action.offset === null) {
//                 return { ...state, timeAdjustOffset: null }
//             }
//             return { ...state, timeAdjustOffset: { ...action.offset } }
//         }
//
//         case 'set_viewing_album': {
//             state.albumId = action.albumId
//             return { ...state }
//         }
//
//         default: {
//             console.error('Do not have handler for dispatch type', action.type)
//             return {
//                 ...state,
//             }
//         }
//     }
// }

export const useKeyDownGallery = () => {
    const menuTargetId = useGalleryStore((state) => state.menuTargetId)
    const blockSearchFocus = useGalleryStore((state) => state.blockSearchFocus)
    const selecting = useGalleryStore((state) => state.selecting)

    const onKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (event.key === 'Shift') {
                useGalleryStore.getState().setHoldingShift(true)
            } else if (event.key === 'Escape') {
                if (useMediaStore.getState().selectedMap.size !== 0) {
                    useMediaStore.getState().clearSelected()
                } else {
                    useGalleryStore.getState().setSelecting(false)
                }
            }
        },
        [blockSearchFocus, menuTargetId, selecting]
    )

    const onKeyUp = useCallback(
        (event: KeyboardEvent) => {
            if (event.key === 'Shift') {
                useGalleryStore.getState().setHoldingShift(false)
            }
        },
        [blockSearchFocus]
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

// export type GalleryContextT = {
//     galleryState: GalleryStateT
//     galleryDispatch: GalleryDispatchT
// }
//
// export const GalleryContext = createContext<GalleryContextT>({
//     galleryState: null,
//     galleryDispatch: null,
// })

export interface GalleryStateT {
    albumsMap: Map<string, AlbumInfo>
    albumsFilter: string[]
    loading: string[]
    newAlbumDialogue: boolean
    blockSearchFocus: boolean
    selecting: boolean
    menuTargetId: string
    imageSize: number
    searchContent: string
    presentingMediaId: string
    presentingMode: PresentType
    timeAdjustOffset: TimeOffset
    hoverIndex: number
    lastSelId: string
    holdingShift: boolean
    albumId: string

    setSelecting: (selecting: boolean) => void
    setHoldingShift: (shift: boolean) => void
    addLoading: (loading: string) => void
    removeLoading: (loading: string) => void
    setTimeOffset: (offset: TimeOffset) => void
    setMenuTarget: (targetId: string) => void
    setPresentationTarget: (mediaId: string, presentMode: PresentType) => void
    setImageSize: (size: number) => void
    setBlockFocus: (block: boolean) => void
}

const GalleryStateControl: StateCreator<
    GalleryStateT,
    [],
    [['zustand/devtools', never]]
> = devtools((set) => ({
    albumsMap: new Map<string, AlbumInfo>(),
    albumsFilter:
        (JSON.parse(localStorage.getItem('albumsFilter')) as string[]) || [],
    loading: [],
    newAlbumDialogue: false,
    blockSearchFocus: false,
    selecting: false,
    menuTargetId: '',
    imageSize: clamp(
        Number(JSON.parse(localStorage.getItem('imageSize'))),
        150,
        500
    ),
    searchContent: '',
    presentingMediaId: '',
    presentingMode: PresentType.None,
    timeAdjustOffset: null,
    hoverIndex: -1,
    lastSelId: '',
    holdingShift: false,
    albumId: '',

    setSelecting: (selecting: boolean) => {
        set({ selecting })
    },

    setHoldingShift: (shift: boolean) => {
        set({ holdingShift: shift })
    },

    addLoading: (loading: string) => {
        set((state) => {
            const newLoading = state.loading.filter((v) => v !== loading)
            newLoading.push(loading)
            return { loading: newLoading }
        })
    },

    removeLoading: (loading: string) => {
        set((state) => {
            const newLoading = state.loading.filter((v) => v !== loading)
            return { loading: newLoading }
        })
    },

    setTimeOffset: (offset: TimeOffset) => {
        set({ timeAdjustOffset: offset })
    },

    setMenuTarget: (targetId: string) => {
        set({ menuTargetId: targetId })
    },

    setPresentationTarget: (mediaId: string, presentMode: PresentType) => {
        set({ presentingMediaId: mediaId, presentingMode: presentMode })
    },

    setImageSize: (size: number) => {
        set({ imageSize: size })
    },

    setBlockFocus: (block: boolean) => {
        set({ blockSearchFocus: block })
    }
}))

export const useGalleryStore = create<GalleryStateT>()(GalleryStateControl)
