import { useDebouncedValue } from '@mantine/hooks'
import React, { useContext, useEffect, useReducer, useRef } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { getAlbums } from '../../Albums/AlbumQuery'
import { Albums } from '../../Albums/Albums'

import HeaderBar from '../../components/HeaderBar'
import Presentation from '../../components/Presentation'
import { UserContext } from '../../Context'

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

const Gallery = () => {
    const [galleryState, galleryDispatch] = useReducer<
        (state: GalleryStateT, action: GalleryAction) => GalleryStateT
    >(galleryReducer, {
        selected: new Map<string, boolean>(),
        albumsMap: new Map<string, AlbumData>(),
        albumsFilter: JSON.parse(localStorage.getItem('albumsFilter')) || [],
        includeRaw: JSON.parse(localStorage.getItem('showRaws')) || false,
        imageSize:
            clamp(JSON.parse(localStorage.getItem('imageSize')), 150, 500) ||
            300,
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
        lastSelIndex: -1,
    })

    const nav = useNavigate()
    const { authHeader, usr } = useContext(UserContext)

    const loc = useLocation()
    const page =
        loc.pathname === '/' || loc.pathname === '/timeline'
            ? 'timeline'
            : 'albums'
    const albumId = useParams()['*']
    const viewportRef: React.Ref<HTMLDivElement> = useRef()

    useKeyDownGallery(galleryState, galleryDispatch)

    useEffect(() => {
        if (usr.isLoggedIn) {
            galleryDispatch({ type: 'remove_loading', loading: 'login' })
        } else if (usr.isLoggedIn === undefined) {
            galleryDispatch({ type: 'add_loading', loading: 'login' })
        } else if (usr.isLoggedIn === false) {
            nav('/login')
        }
    }, [usr])

    useEffect(() => {
        localStorage.setItem(
            'albumsFilter',
            JSON.stringify(galleryState.albumsFilter)
        )
    }, [galleryState.albumsFilter])

    useEffect(() => {
        localStorage.setItem(
            'showRaws',
            JSON.stringify(galleryState.includeRaw)
        )
    }, [galleryState.includeRaw])

    const [bouncedSize] = useDebouncedValue(galleryState.imageSize, 500)
    useEffect(() => {
        localStorage.setItem('imageSize', JSON.stringify(bouncedSize))
    }, [bouncedSize])

    useEffect(() => {
        if (authHeader.Authorization !== '' && page !== 'albums') {
            galleryDispatch({ type: 'add_loading', loading: 'albums' })
            getAlbums(true, authHeader).then((val) => {
                galleryDispatch({ type: 'set_albums', albums: val })
                galleryDispatch({ type: 'remove_loading', loading: 'albums' })
            })
        }
    }, [authHeader, page])

    return (
        <GalleryContext.Provider
            value={{
                galleryState: galleryState,
                galleryDispatch: galleryDispatch,
            }}
        >
            <HeaderBar
                dispatch={galleryDispatch}
                page={'gallery'}
                loading={galleryState.loading}
            />
            <Presentation
                mediaId={
                    galleryState.presentingMode === PresentType.Fullscreen
                        ? galleryState.presentingMediaId
                        : null
                }
                element={null}
                dispatch={galleryDispatch}
            />
            <div className="h-full z-10">
                <div
                    ref={viewportRef}
                    style={{
                        height: 'calc(100% - 80px)',
                        width: '100%',
                        position: 'absolute',
                    }}
                >
                    {page == 'timeline' && <Timeline page={page} />}
                    {page == 'albums' && <Albums selectedAlbum={albumId} />}
                </div>
            </div>
        </GalleryContext.Provider>
    )
}

export default Gallery
