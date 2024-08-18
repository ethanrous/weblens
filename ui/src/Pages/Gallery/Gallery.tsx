import { useDebouncedValue } from '@mantine/hooks'
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
import { Albums } from '../../Albums/Albums'

import HeaderBar from '../../components/HeaderBar'
import Presentation from '../../components/Presentation'

import './galleryStyle.scss'
import { AlbumData, GalleryStateT, PresentType } from '../../types/Types'
import { clamp } from '../../util'
import {
    GalleryAction,
    GalleryContext,
    galleryReducer,
    useKeyDownGallery,
} from './GalleryLogic'
import { Timeline } from './Timeline'
import WeblensButton from '../../components/WeblensButton'
import { IconFilter } from '@tabler/icons-react'
import { MiniAlbumCover } from '../../Albums/AlbumDisplay'
import { useClick, useKeyDown } from '../../components/hooks'
import { useSessionStore } from '../../components/UserInfo'
import { useMediaStore } from '../../Media/MediaStateControl'

export function GalleryFilters() {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const [optionsOpen, setOptionsOpen] = useState(false)
    const [disabledAlbums, setDisabledAlbums] = useState([
        ...galleryState.albumsFilter,
    ])

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
            setShowRaw(raw)
            setShowHidden(hidden)
        },
        [galleryDispatch]
    )

    const closeOptions = useCallback(
        (e) => {
            if (!optionsOpen) {
                return
            }
            e.stopPropagation()
            updateOptions(disabledAlbums, rawOn, hiddenOn)
            setOptionsOpen(false)
        },
        [optionsOpen, disabledAlbums, rawOn, hiddenOn]
    )

    const [dropdownRef, setDropdownRef] = useState(null)

    useClick(closeOptions, dropdownRef)
    useKeyDown('Escape', closeOptions)

    return (
        <div className="flex items-center h-full w-[40px]">
            <WeblensButton
                Left={IconFilter}
                allowRepeat
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
            </div>
        </div>
    )
}

const Gallery = () => {
    const [galleryState, galleryDispatch] = useReducer<
        (state: GalleryStateT, action: GalleryAction) => GalleryStateT
    >(galleryReducer, {
        albumsMap: new Map<string, AlbumData>(),
        albumsFilter: JSON.parse(localStorage.getItem('albumsFilter')) || [],
        imageSize: clamp(
            JSON.parse(localStorage.getItem('imageSize')),
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
        if (server.info.role === 'backup') {
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
                    {page == 'timeline' && <Timeline page={page} />}
                    {page == 'albums' && <Albums selectedAlbum={albumId} />}
                </div>
            </div>
        </GalleryContext.Provider>
    )
}

export default Gallery
