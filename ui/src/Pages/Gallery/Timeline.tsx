import { Indicator, Space, Text } from '@mantine/core'
import { PhotoGallery } from '../../Media/MediaDisplay'
import { WeblensButton } from '../../components/WeblensButton'
import { useNavigate } from 'react-router-dom'
import {
    memo,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from 'react'
import WeblensSlider from '../../components/WeblensSlider'
import { MediaImage } from '../../Media/PhotoContainer'
import { GalleryContext } from './Gallery'
import { IconAlbum, IconFilter, IconFolder } from '@tabler/icons-react'
import { UserContext } from '../../Context'
import WeblensMedia from '../../Media/Media'
import { useClick, useKeyDown } from '../../components/hooks'
import { FetchData } from '../../Media/MediaQuery'

const TimelineControls = () => {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)

    const [optionsOpen, setOptionsOpen] = useState(false)

    const updateOptions = useCallback(
        (disabled: string[], raw: boolean) => {
            galleryDispatch({
                type: 'set_albums_filter',
                albumNames: disabled,
            })
            galleryDispatch({ type: 'set_raw_toggle', raw: raw })
        },
        [galleryDispatch]
    )

    const [disabledAlbums, setDisabledAlbums] = useState([
        ...galleryState.albumsFilter,
    ])
    const [rawOn, setRawOn] = useState(galleryState.includeRaw)

    const albumsOptions = useMemo(() => {
        return Array.from(galleryState.albumsMap.values()).map((a) => {
            const included = !disabledAlbums.includes(a.Id)
            if (!a.CoverMedia) {
                a.CoverMedia = new WeblensMedia({ contentId: a.Cover })
            }
            return (
                <div
                    className="album-selector"
                    key={a.Id}
                    data-included={included}
                    onClick={() =>
                        setDisabledAlbums((p) => {
                            const newP = [...p]
                            if (included) {
                                newP.push(a.Id)
                                return newP
                            } else {
                                newP.splice(newP.indexOf(a.Id))
                                return newP
                            }
                        })
                    }
                >
                    <MediaImage
                        disabled={!included}
                        media={a.CoverMedia}
                        quality="thumbnail"
                        containerStyle={{ borderRadius: 4 }}
                    />
                    <p className="album-selector-title">{a.Name}</p>
                </div>
            )
        })
    }, [galleryState.albumsMap, disabledAlbums])

    const rawClick = useCallback(() => setRawOn(!rawOn), [rawOn, setRawOn])
    const selectClick = useCallback(() => {
        galleryDispatch({
            type: 'set_selecting',
            selecting: !galleryState.selecting,
        })
    }, [galleryDispatch, galleryState.selecting])

    const [dropdownRef, setDropdownRef] = useState(null)

    const closeOptions = useCallback(
        (e) => {
            if (!optionsOpen) {
                return
            }
            e.stopPropagation()
            updateOptions(disabledAlbums, rawOn)
            setOptionsOpen(false)
        },
        [optionsOpen, disabledAlbums, rawOn]
    )

    useClick(closeOptions, dropdownRef)
    useKeyDown('Escape', closeOptions)

    return (
        <div className="flex flex-row items-center grow m-2 h-14 w-[95%]">
            <WeblensSlider
                value={galleryState.imageSize}
                width={200}
                height={35}
                min={150}
                max={500}
                callback={(s) =>
                    galleryDispatch({ type: 'set_image_size', size: s })
                }
            />
            <Space w={20} />
            <div>
                <Indicator
                    color="#4444ff"
                    disabled={
                        !disabledAlbums.length && !galleryState.includeRaw
                    }
                    zIndex={3}
                >
                    <IconFilter
                        onClick={() => setOptionsOpen((p) => !p)}
                        style={{ cursor: 'pointer' }}
                    />
                </Indicator>
            </div>
            <div
                className="options-dropdown"
                data-open={optionsOpen}
                ref={setDropdownRef}
            >
                <div className="flex flex-col items-center p-2">
                    <div style={{ paddingBottom: 10 }}>
                        <Text fw={600}>Gallery Filters</Text>
                    </div>

                    <Space h={10} />
                    <WeblensButton
                        label="Show RAWs"
                        squareSize={40}
                        allowRepeat
                        toggleOn={rawOn}
                        onClick={rawClick}
                    />
                    <Space h={10} />
                    <div className="grid grid-cols-2 gap-2 max-h-[500px] overflow-y-scroll no-scrollbar">
                        {albumsOptions}
                    </div>
                </div>
            </div>
            <div className="flex grow w-0 justify-end">
                <WeblensButton
                    label="Select"
                    allowRepeat
                    squareSize={40}
                    centerContent
                    toggleOn={galleryState.selecting}
                    onClick={selectClick}
                    disabled={galleryState.mediaMap.size === 0}
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
                <Text c="white" fw={700} size="31px">
                    No media to display
                </Text>
                <Text c="white">Upload files then add them to an album</Text>
                <div className="h-max w-full gap-2">
                    <WeblensButton
                        squareSize={48}
                        fillWidth
                        label="FileBrowser"
                        Left={IconFolder}
                        // centerContent
                        onClick={() => nav('/files')}
                    />
                    <WeblensButton
                        squareSize={48}
                        fillWidth
                        label="Albums"
                        Left={IconAlbum}
                        // centerContent
                        onClick={() => nav('/albums')}
                    />
                </div>
            </div>
        </div>
    )
}

export const Timeline = memo(
    ({ page }: { page: string }) => {
        const { galleryState, galleryDispatch } = useContext(GalleryContext)
        const { authHeader } = useContext(UserContext)

        useEffect(() => {
            if (!galleryState) {
                return
            }

            galleryDispatch({ type: 'add_loading', loading: 'media' })
            FetchData(galleryState, galleryDispatch, authHeader).then(() =>
                galleryDispatch({ type: 'remove_loading', loading: 'media' })
            )
        }, [
            galleryState?.includeRaw,
            galleryState?.albumsFilter,
            galleryState?.albumsMap,
            page,
            authHeader,
        ])

        const medias = useMemo(() => {
            if (!galleryState) {
                return []
            }

            return Array.from(galleryState.mediaMap.values())
                .filter((m) => {
                    if (galleryState.searchContent === '') {
                        return true
                    }

                    return m.MatchRecogTag(galleryState.searchContent)
                })
                .reverse()
        }, [galleryState?.mediaMap, galleryState?.searchContent])

        if (galleryState.loading.includes('media')) {
            return null
        }

        return (
            <div className="flex flex-col items-center">
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
