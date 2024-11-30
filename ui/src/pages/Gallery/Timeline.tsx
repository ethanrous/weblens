import { IconLayersIntersect } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import MediaApi from '@weblens/api/MediaApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { GalleryFilters } from '@weblens/pages/Gallery/Gallery'
import WeblensMedia from '@weblens/types/media/Media'
import { PhotoGallery } from '@weblens/types/media/MediaDisplay'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'

import { useGalleryStore } from './GalleryLogic'

function TimelineControls() {
    const selecting = useGalleryStore((state) => state.selecting)
    const imageSize = useGalleryStore((state) => state.imageSize)
    const setSelecting = useGalleryStore((state) => state.setSelecting)
    const setImageSize = useGalleryStore((state) => state.setImageSize)

    const hasMedia = useMediaStore((state) => state.mediaMap.size !== 0)

    return (
        <div className="timeline-controls">
            <div className="relative h-10 w-56 shrink-0">
                <WeblensProgress
                    height={40}
                    value={((imageSize - 150) / 350) * 100}
                    disabled={selecting}
                    seekCallback={(s) => {
                        if (s === 0) {
                            s = 1
                        }
                        const newSize = (s / 100) * 350 + 150
                        setImageSize(newSize)
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
                    toggleOn={selecting}
                    onClick={() => setSelecting(!selecting)}
                    disabled={!hasMedia}
                />
            </div>
        </div>
    )
}

export function Timeline() {
    // const loading = useGalleryStore((state) => state.loading)
    // const albumsFilter = useGalleryStore((state) => state.albumsFilter)
    // const albumsMap = useGalleryStore((state) => state.albumsMap)
    // const addLoading = useGalleryStore((state) => state.addLoading)
    // const removeLoading = useGalleryStore((state) => state.removeLoading)

    const showRaw = useMediaStore((state) => state.showRaw)
    const showHidden = useMediaStore((state) => state.showHidden)
    // const medias = useMediaStore((state) => [...state.mediaMap.values()])
    const addMedias = useMediaStore((state) => state.addMedias)

    const {
        data: medias,
        isLoading,
        error,
    } = useQuery<WeblensMedia[]>({
        queryKey: ['media', showRaw, showHidden],
        initialData: [],
        queryFn: async () => {
            console.log('Getting media')
            const res = await MediaApi.getMedia(
                showRaw,
                showHidden,
                undefined,
                0,
                10000
            ).then((res) => {
                return res.data.Media.map((info) => new WeblensMedia(info))
            })

            addMedias(res)
            return res
        },
    })

    // useEffect(() => {
    //     if (loading.includes('media')) {
    //         return
    //     }
    //     addLoading('media')
    //     MediaApi.getMedia(showRaw, showHidden, undefined, 0, 10000)
    //         .then((res) => {
    //             const medias = res.data.Media.map((info) => {
    //                 return new WeblensMedia(info)
    //             })
    //
    //             addMedias(medias)
    //             removeLoading('media')
    //         })
    //         .catch((err) => {
    //             console.error('Failed to get media', err)
    //         })
    // }, [showRaw, showHidden, albumsFilter, albumsMap])

    return (
        <div className="flex flex-col items-center h-1/2 w-full relative grow">
            <TimelineControls />
            <PhotoGallery medias={medias} loading={isLoading} error={error} />
        </div>
    )
}
