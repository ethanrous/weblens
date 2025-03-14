import HeaderBar from '@weblens/components/HeaderBar'
import Presentation from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import { useClick, useKeyDown } from '@weblens/lib/hooks'
import { PresentType } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import React, { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { useGalleryStore, useKeyDownGallery } from './GalleryLogic'
import { Timeline } from './Timeline'
import './galleryStyle.scss'

export function GalleryFilters() {
    // const selecting = useGalleryStore((state) => state.selecting)

    const [optionsOpen, setOptionsOpen] = useState(false)
    const clearMedias = useMediaStore((state) => state.clear)

    const showRaw = useMediaStore((state) => state.showRaw)
    const showHidden = useMediaStore((state) => state.showHidden)

    const setShowRaw = useMediaStore((state) => state.setShowingRaw)
    const setShowHidden = useMediaStore((state) => state.setShowingHidden)

    const [rawOn, setRawOn] = useState(showRaw)
    const [hiddenOn, setHiddenOn] = useState(showHidden)

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
        <div className="relative flex h-max w-[40px] flex-col items-center">
            {/* <WeblensButton */}
            {/*     Left={IconFilter} */}
            {/*     allowRepeat */}
            {/*     tooltip="Gallery Filters" */}
            {/*     onClick={() => setOptionsOpen((p) => !p)} */}
            {/*     disabled={selecting} */}
            {/*     toggleOn={disabledAlbums.length !== 0 || showRaw} */}
            {/* /> */}
            <div
                className="wl-floating-card pointer-events-auto absolute z-50 mt-12 flex w-max flex-col items-center rounded-md shadow-xl transition data-[open=false]:pointer-events-none data-[open=false]:scale-90 data-[open=false]:opacity-0"
                data-open={optionsOpen}
                ref={setDropdownRef}
            >
                <div className="flex flex-col items-center gap-4 p-2">
                    <p className="text-lg font-semibold">Gallery Filters</p>
                    <div className="flex h-max w-full gap-3">
                        <WeblensButton
                            label="Show RAWs"
                            flavor={rawOn ? 'default' : 'outline'}
                            onClick={() => setRawOn((raw) => !raw)}
                        />
                        <WeblensButton
                            label="Show Hidden"
                            centerContent
                            flavor={hiddenOn ? 'default' : 'outline'}
                            onClick={() => {
                                setHiddenOn((hidden) => !hidden)
                            }}
                        />
                    </div>
                </div>
                <div className="flex w-full justify-end">
                    <WeblensButton
                        label="Save"
                        centerContent
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

    const imageSize = useGalleryStore((state) => state.imageSize)
    const presentingMode = useGalleryStore((state) => state.presentingMode)
    const presentingMediaId = useGalleryStore(
        (state) => state.presentingMediaId
    )
    const setPresentationTarget = useGalleryStore(
        (state) => state.setPresentationTarget
    )
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
        localStorage.setItem('imageSize', JSON.stringify(imageSize))
    }, [imageSize])

    return (
        <div className="relative flex h-screen w-screen flex-col">
            <HeaderBar />
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
                className="relative flex h-[50%] w-full shrink-0 grow flex-col"
            >
                {page == 'timeline' && <Timeline />}
                {/* {page == 'albums' && <Albums selectedAlbumId={albumId} />} */}
            </div>
        </div>
    )
}

export default Gallery
