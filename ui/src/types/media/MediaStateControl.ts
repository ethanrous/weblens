import Media from '@weblens/types/media/Media'
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export interface MediaStateT {
    mediaMap: Map<string, Media>
    selectedMap: Map<string, Media>

    hoverId: string
    lastSelectedId: string

    showRaw: boolean
    showHidden: boolean

    mediaTypeMap: any

    addMedias: (medias: Media[]) => void
    setShowingRaw: (showRaw: boolean) => void
    setShowingHidden: (showHidden: boolean) => void
    hideMedias: (mediaIds: string[], hidden: boolean) => void
    clearSelected: () => void
    setHovering: (hoveringId: string) => void
    setSelected: (mediaId: string, selected: boolean) => void
    setLiked: (mediaId: string, likedBy: string) => void
    setTypeMap: (typeMap: any) => void
    getTypeMap: () => any
    clear: () => void
}

export const useMediaStore = create<MediaStateT>()(
    devtools((set, get) => ({
        mediaMap: new Map<string, Media>(),
        selectedMap: new Map<string, Media>(),
        showRaw: JSON.parse(localStorage.getItem('showRaws')) || false,
        showHidden: JSON.parse(localStorage.getItem('showHidden')) || false,
        hoverId: '',
        lastSelectedId: '',
        mediaTypeMap: null,

        addMedias: (medias) => {
            set((state) => {
                for (const media of medias) {
                    state.mediaMap.set(media.Id(), media)
                }

                return {
                    mediaMap: new Map(state.mediaMap),
                }
            })
        },

        setShowingRaw: (showRaw: boolean) => {
            localStorage.setItem('showRaws', String(showRaw))
            set({ showRaw: showRaw })
        },

        setShowingHidden: (showHidden: boolean) => {
            localStorage.setItem('showHidden', String(showHidden))
            set({ showHidden: showHidden })
        },

        hideMedias: (mediaIds: string[], hidden: boolean) => {
            const newMap = get().mediaMap
            for (const mediaId of mediaIds) {
                const m = newMap.get(mediaId)
                if (!m) {
                    console.error('trying to hide unknown mediaId', mediaId)
                    continue
                }
                m.SetHidden(hidden)
            }

            set({
                selectedMap: new Map<string, Media>(),
                mediaMap: new Map(newMap),
            })
        },

        clearSelected: () => {
            set((state) => {
                state.selectedMap.forEach((m) => m.SetSelected(false))
                return { selectedMap: new Map<string, Media>() }
            })
        },

        setHovering: (hoverId: string) => {
            set({ hoverId: hoverId })
        },

        setSelected: (mediaId: string, selected: boolean) => {
            const media = get().mediaMap.get(mediaId)
            if (!media) {
                console.error(
                    'Cannot find media id trying to select media',
                    mediaId
                )
                return
            }
            if (media.IsSelected() === selected) {
                return
            }

            media.SetSelected(selected)

            if (selected) {
                get().selectedMap.set(mediaId, media)
            } else {
                get().selectedMap.delete(mediaId)
            }

            set((state) => {
                return {
                    selectedMap: new Map(state.selectedMap),
                    mediaMap: new Map(state.mediaMap),
                }
            })
        },

        setLiked: (mediaId: string, likedBy: string) => {
            const media = get().mediaMap.get(mediaId)
            if (!media) {
                console.error(
                    'Cannot find mediaId trying to like media',
                    mediaId
                )
                return
            }
            media.SetLikedBy(likedBy)
            console.log('LIKED', media.GetLikedBy())

            set((state) => ({ mediaMap: new Map(state.mediaMap) }))
        },

        clear: () => {
            set({
                selectedMap: new Map<string, Media>(),
                mediaMap: new Map<string, Media>(),
                hoverId: '',
                lastSelectedId: '',
            })
        },

        setTypeMap: (typeMap: any) => {
            set({ mediaTypeMap: typeMap })
        },

        getTypeMap: () => {
            return get().mediaTypeMap
        },
    }))
)
