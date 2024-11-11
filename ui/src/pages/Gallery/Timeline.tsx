import { IconFolder, IconLayersIntersect } from '@tabler/icons-react'
import MediaApi from '@weblens/api/MediaApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { GalleryFilters } from '@weblens/pages/Gallery/Gallery'
import { GalleryContext } from '@weblens/pages/Gallery/GalleryLogic'
import WeblensMedia from '@weblens/types/media/Media'
import { PhotoGallery } from '@weblens/types/media/MediaDisplay'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { memo, useCallback, useContext, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'

const TimelineControls = () => {
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

const NoMediaDisplay = () => {
    const nav = useNavigate()
    return (
        <div className="flex flex-col items-center w-full">
            <div className="flex flex-col items-center mt-20 gap-2 w-[300px]">
                <h2 className="font-bold text-3xl select-none">
                    No media to display
                </h2>
                <p className="select-none">
                    Upload files or adjust the filters
                </p>
                <div className="h-max w-full gap-2">
                    <WeblensButton
                        squareSize={48}
                        fillWidth
                        label="FileBrowser"
                        Left={IconFolder}
                        onClick={() => nav('/files')}
                    />
                </div>
            </div>
        </div>
    )
}

export const Timeline = memo(
    ({ page }: { page: string }) => {
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
            MediaApi.getMedia(showRaw, showHidden).then((res) => {
                const medias = res.data.Media.map((info) => {
                    return new WeblensMedia(info)
                })

                addMedias(medias)
                galleryDispatch({
                    type: 'remove_loading',
                    loading: 'media',
                })
            })
        }, [
            showRaw,
            showHidden,
            galleryState?.albumsFilter,
            galleryState?.albumsMap,
            page,
        ])

        if (galleryState.loading.includes('media')) {
            return null
        }

        return (
            <div className="flex flex-col items-center h-1/2 w-full relative grow">
                <TimelineControls />
                {medias.length === 0 && <NoMediaDisplay />}
                {medias.length !== 0 && <PhotoGallery medias={medias} />}
            </div>
        )
    },
    (prev, next) => {
        return prev.page === next.page
    }
)
