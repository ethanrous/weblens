import { useDebouncedValue } from '@mantine/hooks'
import { IconFilter } from '@tabler/icons-react'
import { AlbumInfo } from '@weblens/api/swag'
import HeaderBar from '@weblens/components/HeaderBar'
import Presentation from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useClick, useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { GalleryStateT, PresentType } from '@weblens/types/Types'
import { MiniAlbumCover } from '@weblens/types/albums/AlbumDisplay'
import { Albums } from '@weblens/types/albums/Albums'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { clamp } from '@weblens/util'
import React, {
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useReducer,
    useRef,
    useState,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import {
    GalleryAction,
    GalleryContext,
    galleryReducer,
    useKeyDownGallery,
} from './GalleryLogic'
import { Timeline } from './Timeline'
import './galleryStyle.scss'

export function GalleryFilters() {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const [optionsOpen, setOptionsOpen] = useState(false)
    const [disabledAlbums] = useState([...galleryState.albumsFilter])
    const clearMedias = useMediaStore((state) => state.clear)

    const showRaw = useMediaStore((state) => state.showRaw)
    const showHidden = useMediaStore((state) => state.showHidden)

    const setShowRaw = useMediaStore((state) => state.setShowingRaw)
    const setShowHidden = useMediaStore((state) => state.setShowingHidden)

    const [rawOn, setRawOn] = useState(showRaw)
    const [hiddenOn, setHiddenOn] = useState(showHidden)

    const albumsOptions = useMemo(() => {
        return Array.from(galleryState.albumsMap.values()).map((a) => {
            return <MiniAlbumCover key={a.id} album={a} />
        })
    }, [galleryState.albumsMap, disabledAlbums])

    const updateOptions = useCallback(
        (disabled: string[], raw: boolean, hidden: boolean) => {
            galleryDispatch({
                type: 'set_albums_filter',
                albumNames: disabled,
            })
            clearMedias()
            setShowRaw(raw)
            setShowHidden(hidden)
        },
        [galleryDispatch]
    )

    const [dropdownRef, setDropdownRef] = useState<HTMLDivElement>(null)

    useClick((e) => {
        if (optionsOpen) {
            e.stopPropagation()
            setOptionsOpen(false)
        }
    }, dropdownRef)
    useKeyDown('Escape', (e) => {
        if (optionsOpen) {
            e.stopPropagation()
            setOptionsOpen(false)
        }
    })

    return (
        <div className="flex flex-col items-center h-full w-[40px]">
            <WeblensButton
                Left={IconFilter}
                allowRepeat
                tooltip="Gallery Filters"
                onClick={() => setOptionsOpen((p) => !p)}
                disabled={galleryState.selecting}
                toggleOn={disabledAlbums.length !== 0 || showRaw}
            />
            <div
                className="options-dropdown"
                data-open={optionsOpen}
                ref={setDropdownRef}
            >
                <div className="flex flex-col items-center p-2 gap-4">
                    <p className="font-semibold text-lg">Gallery Filters</p>
                    <div className="flex gap-3">
                        <WeblensButton
                            label="Show RAWs"
                            squareSize={40}
                            allowRepeat
                            toggleOn={rawOn}
                            onClick={() => setRawOn((raw) => !raw)}
                        />
                        <WeblensButton
                            label="Show Hidden"
                            squareSize={40}
                            allowRepeat
                            toggleOn={hiddenOn}
                            onClick={() => {
                                setHiddenOn((hidden) => !hidden)
                            }}
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-2 max-h-[500px] overflow-y-scroll no-scrollbar">
                        {albumsOptions}
                    </div>
                </div>
                <div className="flex justify-end w-full">
                    <WeblensButton
                        label="Save"
                        disabled={rawOn === showRaw && hiddenOn === showHidden}
                        onClick={(e) => {
                            if (optionsOpen) {
                                e.stopPropagation()
                                updateOptions(disabledAlbums, rawOn, hiddenOn)
                                setOptionsOpen(false)
                            }
                        }}
                    />
                </div>
            </div>
        </div>
    )
}

const Gallery = () => {
    const [galleryState, galleryDispatch] = useReducer<
        (state: GalleryStateT, action: GalleryAction) => GalleryStateT
    >(galleryReducer, {
        albumsMap: new Map<string, AlbumInfo>(),
        albumsFilter:
            (JSON.parse(localStorage.getItem('albumsFilter')) as string[]) ||
            [],
        imageSize: clamp(
            Number(JSON.parse(localStorage.getItem('imageSize'))),
            150,
            500
        ),
        presentingMediaId: '',
        presentingMode: PresentType.None,
        loading: [],
        newAlbumDialogue: false,
        blockSearchFocus: false,
        selecting: false,
        searchContent: '',
        menuTargetId: '',
        timeAdjustOffset: null,
        holdingShift: false,
        hoverIndex: -1,
        lastSelId: '',
        albumId: '',
    })

    const nav = useNavigate()
    const server = useSessionStore((state) => state.server)

    const loc = useLocation()
    const page =
        loc.pathname === '/' || loc.pathname === '/timeline'
            ? 'timeline'
            : 'albums'
    const albumId = useParams()['*']
    const viewportRef: React.Ref<HTMLDivElement> = useRef()

    useKeyDownGallery(galleryState, galleryDispatch)

    useEffect(() => {
        if (!server) {
            return
        }
        if (server.role === 'backup') {
            nav('/files/home')
        }
    }, [server])

    useEffect(() => {
        localStorage.setItem(
            'albumsFilter',
            JSON.stringify(galleryState.albumsFilter)
        )
    }, [galleryState.albumsFilter])

    const [bouncedSize] = useDebouncedValue(galleryState.imageSize, 500)
    useEffect(() => {
        localStorage.setItem('imageSize', JSON.stringify(bouncedSize))
    }, [bouncedSize])

    useEffect(() => {
        galleryDispatch({ type: 'set_viewing_album', albumId: albumId })
    }, [albumId])

    return (
        <GalleryContext.Provider
            value={{
                galleryState: galleryState,
                galleryDispatch: galleryDispatch,
            }}
        >
            <div className="flex flex-col h-screen w-screen relative">
                <HeaderBar
                    page={'gallery'}
                    loading={galleryState.loading}
                    setBlockFocus={(block: boolean) =>
                        galleryDispatch({
                            type: 'set_block_focus',
                            block: block,
                        })
                    }
                />
                <Presentation
                    mediaId={
                        galleryState.presentingMode === PresentType.Fullscreen
                            ? galleryState.presentingMediaId
                            : null
                    }
                    element={null}
                    dispatch={{
                        setPresentationTarget: (targetId: string) =>
                            galleryDispatch({
                                type: 'set_presentation',
                                mediaId: targetId,
                            }),
                    }}
                />

                <div
                    ref={viewportRef}
                    className="flex flex-col h-[50%] w-full shrink-0 relative grow"
                >
                    {page == 'timeline' && <Timeline />}
                    {page == 'albums' && <Albums selectedAlbumId={albumId} />}
                </div>
            </div>
        </GalleryContext.Provider>
    )
}

export default Gallery
