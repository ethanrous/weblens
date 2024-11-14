import { IconLayersIntersect } from '@tabler/icons-react'
import MediaApi from '@weblens/api/MediaApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { GalleryFilters } from '@weblens/pages/Gallery/Gallery'
import { GalleryContext } from '@weblens/pages/Gallery/GalleryLogic'
import WeblensMedia from '@weblens/types/media/Media'
import { PhotoGallery } from '@weblens/types/media/MediaDisplay'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { useCallback, useContext, useEffect } from 'react'

function TimelineControls() {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const hasMedia = useMediaStore((state) => state.mediaMap.size !== 0)

    const selectClick = useCallback(() => {
        galleryDispatch({
            type: 'set_selecting',
            selecting: !galleryState.selecting,
        })
    }, [galleryDispatch, galleryState.selecting])

    return (
        <div className="timeline-controls">
            <div className="relative h-10 w-56 shrink-0">
                <WeblensProgress
                    height={40}
                    value={((galleryState.imageSize - 150) / 350) * 100}
                    disabled={galleryState.selecting}
                    seekCallback={(s) => {
                        if (s === 0) {
                            s = 1
                        }
                        galleryDispatch({
                            type: 'set_image_size',
                            size: (s / 100) * 350 + 150,
                        })
                    }}
                />
            </div>

            <GalleryFilters />

            <div className="flex grow w-0 justify-end">
                <WeblensButton
                    label="Select"
                    allowRepeat
                    squareSize={40}
                    centerContent
                    Left={IconLayersIntersect}
                    toggleOn={galleryState.selecting}
                    onClick={selectClick}
                    disabled={!hasMedia}
                />
            </div>
        </div>
    )
}

export function Timeline() {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const showRaw = useMediaStore((state) => state.showRaw)
    const showHidden = useMediaStore((state) => state.showHidden)
    const medias = useMediaStore((state) => [...state.mediaMap.values()])
    const addMedias = useMediaStore((state) => state.addMedias)

    useEffect(() => {
        if (!galleryState || galleryState.loading.includes('media')) {
            return
        }

        galleryDispatch({ type: 'add_loading', loading: 'media' })
        MediaApi.getMedia(showRaw, showHidden, undefined, 0, 10000)
            .then((res) => {
                const medias = res.data.Media.map((info) => {
                    return new WeblensMedia(info)
                })

                addMedias(medias)
                galleryDispatch({
                    type: 'remove_loading',
                    loading: 'media',
                })
            })
            .catch((err) => {
                console.error('Failed to get media', err)
            })
    }, [
        showRaw,
        showHidden,
        galleryState?.albumsFilter,
        galleryState?.albumsMap,
    ])

    // if (galleryState.loading.includes('media')) {
    //     return null
    // }

    return (
        <div className="flex flex-col items-center h-1/2 w-full relative grow">
            <TimelineControls />
            <PhotoGallery medias={medias} />
        </div>
    )
}
