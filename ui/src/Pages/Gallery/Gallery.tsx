import { Box } from '@mantine/core'
import React, {
    createContext,
    useContext,
    useEffect,
    useReducer,
    useRef,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import HeaderBar from '../../components/HeaderBar'
import Presentation from '../../components/Presentation'
import { GalleryAction, mediaReducer, useKeyDownGallery } from './GalleryLogic'
import {
    AlbumData,
    GalleryDispatchT,
    GalleryStateT,
    PresentType,
    UserInfoT,
} from '../../types/Types'
import { UserContext } from '../../Context'
import { Albums } from '../../Albums/Albums'
import { useDebouncedValue } from '@mantine/hooks'
import { clamp } from '../../util'
import WeblensMedia from '../../Media/Media'

import './galleryStyle.scss'
import { Timeline } from './Timeline'
import { getAlbums } from '../../Albums/AlbumQuery'

export type GalleryContextT = {
    galleryState: GalleryStateT
    galleryDispatch: GalleryDispatchT
}

export const GalleryContext = createContext<GalleryContextT>({
    galleryState: null,
    galleryDispatch: null,
})

const Gallery = () => {
    const [galleryState, galleryDispatch] = useReducer<
        (state: GalleryStateT, action: GalleryAction) => GalleryStateT
    >(mediaReducer, {
        mediaMap: new Map<string, WeblensMedia>(),
        selected: new Map<string, boolean>(),
        albumsMap: new Map<string, AlbumData>(),
        albumsFilter: JSON.parse(localStorage.getItem('albumsFilter')) || [],
        includeRaw: JSON.parse(localStorage.getItem('showRaws')) || false,
        imageSize:
            clamp(JSON.parse(localStorage.getItem('imageSize')), 150, 500) ||
            300,
        presentingMedia: null,
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
    const { authHeader, usr }: { authHeader; usr: UserInfoT } =
        useContext(UserContext)

    const loc = useLocation()
    const page =
        loc.pathname === '/' || loc.pathname === '/timeline'
            ? 'timeline'
            : 'albums'
    const albumId = useParams()['*']
    const viewportRef: React.Ref<HTMLDivElement> = useRef()

    const searchRef = useRef()
    useKeyDownGallery(searchRef, galleryState, galleryDispatch)

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
            getAlbums(authHeader).then((val) => {
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
                itemId={galleryState.presentingMedia?.Id()}
                mediaData={
                    galleryState.presentingMode === PresentType.Fullscreen
                        ? galleryState.presentingMedia
                        : null
                }
                element={null}
                dispatch={galleryDispatch}
            />
            <div className="h-full z-10">
                <Box
                    ref={viewportRef}
                    style={{
                        height: 'calc(100% - 80px)',
                        width: '100%',
                        position: 'absolute',
                    }}
                >
                    {page == 'timeline' && <Timeline page={page} />}
                    {page == 'albums' && <Albums selectedAlbum={albumId} />}
                </Box>
            </div>
        </GalleryContext.Provider>
    )
}

export default Gallery
