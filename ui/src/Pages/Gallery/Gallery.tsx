import { Box, Button, Modal, Space, Tabs, TextInput } from '@mantine/core'
import {
    useEffect,
    useReducer,
    useRef,
    useContext,
    useState,
    createContext,
    useCallback,
} from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { IconFilter, IconPlus } from '@tabler/icons-react'

import HeaderBar from '../../components/HeaderBar'
import Presentation from '../../components/Presentation'
import { mediaReducer, useKeyDownGallery, GalleryAction } from './GalleryLogic'
import { CreateAlbum, FetchData, GetAlbums } from '../../api/GalleryApi'
import {
    AlbumData,
    GalleryDispatchT,
    GalleryStateT,
    PresentType,
    UserContextT,
    UserInfoT,
} from '../../types/Types'
import { UserContext } from '../../Context'
import { RowBox } from '../FileBrowser/FileBrowserStyles'
import { Albums } from './Albums'
import { WeblensButton } from '../../components/WeblensButton'
import { useDebouncedValue } from '@mantine/hooks'
import { clamp } from '../../util'

import WeblensSlider from '../../components/WeblensSlider'
import WeblensMedia from '../../classes/Media'

import './galleryStyle.scss'
import { Timeline } from './Timeline'

export type GalleryContextT = {
    galleryState: GalleryStateT
    galleryDispatch: GalleryDispatchT
}

export const GalleryContext = createContext<GalleryContextT>({
    galleryState: null,
    galleryDispatch: null,
})

const Gallery = () => {
    const [mediaState, dispatch] = useReducer<
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
    useKeyDownGallery(searchRef, mediaState, dispatch)

    useEffect(() => {
        if (usr.isLoggedIn) {
            dispatch({ type: 'remove_loading', loading: 'login' })
        } else if (usr.isLoggedIn === undefined) {
            dispatch({ type: 'add_loading', loading: 'login' })
        } else if (usr.isLoggedIn === false) {
            nav('/login')
        }
    }, [usr])

    useEffect(() => {
        localStorage.setItem(
            'albumsFilter',
            JSON.stringify(mediaState.albumsFilter)
        )
    }, [mediaState.albumsFilter])

    useEffect(() => {
        localStorage.setItem('showRaws', JSON.stringify(mediaState.includeRaw))
    }, [mediaState.includeRaw])

    const [bouncedSize] = useDebouncedValue(mediaState.imageSize, 500)
    useEffect(() => {
        localStorage.setItem('imageSize', JSON.stringify(bouncedSize))
    }, [bouncedSize])

    useEffect(() => {
        if (authHeader.Authorization !== '' && page !== 'albums') {
            dispatch({ type: 'add_loading', loading: 'albums' })
            GetAlbums(authHeader).then((val) => {
                dispatch({ type: 'set_albums', albums: val })
                dispatch({ type: 'remove_loading', loading: 'albums' })
            })
        }
    }, [authHeader, page])

    return (
        <GalleryContext.Provider
            value={{ galleryState: mediaState, galleryDispatch: dispatch }}
        >
            <HeaderBar
                dispatch={dispatch}
                page={'gallery'}
                loading={mediaState.loading}
            />
            <Presentation
                itemId={mediaState.presentingMedia?.Id()}
                mediaData={
                    mediaState.presentingMode === PresentType.Fullscreen
                        ? mediaState.presentingMedia
                        : null
                }
                element={null}
                dispatch={dispatch}
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
