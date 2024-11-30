import { useDebouncedValue } from '@mantine/hooks'
import { IconFilter } from '@tabler/icons-react'
import HeaderBar from '@weblens/components/HeaderBar'
import Presentation from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useClick, useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { PresentType } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import React, { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { useGalleryStore, useKeyDownGallery } from './GalleryLogic'
import { Timeline } from './Timeline'
import './galleryStyle.scss'

export function GalleryFilters() {
    const albumsFilter = useGalleryStore((state) => state.albumsFilter)
    const selecting = useGalleryStore((state) => state.selecting)

    const [optionsOpen, setOptionsOpen] = useState(false)
    const [disabledAlbums] = useState([...albumsFilter])
    const clearMedias = useMediaStore((state) => state.clear)

    const showRaw = useMediaStore((state) => state.showRaw)
    const showHidden = useMediaStore((state) => state.showHidden)

    const setShowRaw = useMediaStore((state) => state.setShowingRaw)
    const setShowHidden = useMediaStore((state) => state.setShowingHidden)

    const [rawOn, setRawOn] = useState(showRaw)
    const [hiddenOn, setHiddenOn] = useState(showHidden)

    // const albumsOptions = useMemo(() => {
    //     return Array.from(albumsMap.values()).map((a) => {
    //         return <MiniAlbumCover key={a.id} album={a} />
    //     })
    // }, [galleryState.albumsMap, disabledAlbums])

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
                disabled={selecting}
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

                    {/* <div className="grid grid-cols-2 gap-2 max-h-[500px] overflow-y-scroll no-scrollbar"> */}
                    {/*     {albumsOptions} */}
                    {/* </div> */}
                </div>
                <div className="flex justify-end w-full">
                    <WeblensButton
                        label="Save"
                        disabled={rawOn === showRaw && hiddenOn === showHidden}
                        onClick={(e) => {
                            if (optionsOpen) {
                                e.stopPropagation()
                                clearMedias()
                                setShowRaw(rawOn)
                                setShowHidden(hiddenOn)
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
    const nav = useNavigate()

    const albumsFilter = useGalleryStore((state) => state.albumsFilter)
    const imageSize = useGalleryStore((state) => state.imageSize)
    const presentingMode = useGalleryStore((state) => state.presentingMode)
    const presentingMediaId = useGalleryStore(
        (state) => state.presentingMediaId
    )
    const setPresentationTarget = useGalleryStore(
        (state) => state.setPresentationTarget
    )
    const setBlockFocus = useGalleryStore((state) => state.setBlockFocus)

    const server = useSessionStore((state) => state.server)

    const loc = useLocation()
    const page =
        loc.pathname === '/' || loc.pathname === '/timeline'
            ? 'timeline'
            : 'albums'
    // const albumId = useParams()['*']
    const viewportRef: React.Ref<HTMLDivElement> = useRef()

    useKeyDownGallery()

    useEffect(() => {
        if (!server) {
            return
        }
        if (server.role === 'backup') {
            nav('/files/home')
        }
    }, [server])

    useEffect(() => {
        localStorage.setItem('albumsFilter', JSON.stringify(albumsFilter))
    }, [albumsFilter])

    const [bouncedSize] = useDebouncedValue(imageSize, 500)
    useEffect(() => {
        localStorage.setItem('imageSize', JSON.stringify(bouncedSize))
    }, [bouncedSize])

    return (
        <div className="flex flex-col h-screen w-screen relative">
            <HeaderBar
                page={'gallery'}
                setBlockFocus={(block: boolean) => setBlockFocus(block)}
            />
            <Presentation
                mediaId={
                    presentingMode === PresentType.Fullscreen
                        ? presentingMediaId
                        : null
                }
                element={null}
                setTarget={(targetId: string) =>
                    setPresentationTarget(targetId, PresentType.Fullscreen)
                }
            />

            <div
                ref={viewportRef}
                className="flex flex-col h-[50%] w-full shrink-0 relative grow"
            >
                {page == 'timeline' && <Timeline />}
                {/* {page == 'albums' && <Albums selectedAlbumId={albumId} />} */}
            </div>
        </div>
    )
}

export default Gallery
