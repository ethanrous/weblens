import { useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { notifications } from '@mantine/notifications'
import { Divider, Space, Text } from '@mantine/core'
import {
    IconArrowLeft,
    IconFolder,
    IconLibraryPlus,
    IconPlus,
} from '@tabler/icons-react'

import { AlbumData, UserContextT } from '../types/Types'
import WeblensMedia from '../Media/Media'
import { UserContext } from '../Context'
import { PhotoGallery } from '../Media/MediaDisplay'
import NotFound from '../components/NotFound'
import { GalleryContext, GalleryContextT } from '../Pages/Gallery/Gallery'
import { useMediaType } from '../components/hooks'
import { AlbumScroller } from './AlbumDisplay'
import { WeblensButton } from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'
import WeblensSlider from '../components/WeblensSlider'
import { createAlbum, GetAlbumMedia, getAlbums } from './AlbumQuery'

function AlbumNoContent({
    albumData,
}: {
    albumData: {
        albumMeta: AlbumData
        media: WeblensMedia[]
    }
}) {
    const nav = useNavigate()
    return (
        <div className="flex flex-col w-full items-center">
            <Text
                size={'75px'}
                fw={900}
                variant="gradient"
                style={{
                    display: 'flex',
                    justifyContent: 'center',
                    userSelect: 'none',
                    lineHeight: 1.1,
                }}
            >
                {albumData.albumMeta.Name}
            </Text>
            <div className="flex flex-col pt-40 w-max items-center">
                {albumData.albumMeta.Medias.length !== 0 && (
                    <div className="flex flex-col items-center">
                        <p className="font-extrabold text-3xl">
                            No media in current filters
                        </p>
                        <Space h={5} />
                        <p className="font-medium text-xl">
                            Adjust the filters
                        </p>
                        <Space h={5} />
                        <Divider label="or" mx={30} />
                    </div>
                )}
                {albumData.albumMeta.Medias.length === 0 && (
                    <p className="font-extrabold text-3xl">
                        This album has no media
                    </p>
                )}
                <Space h={10} />
                <WeblensButton
                    squareSize={40}
                    centerContent
                    label="FileBrowser"
                    Left={IconFolder}
                    onClick={() => nav('/files/home')}
                />
            </div>
        </div>
    )
}

function AlbumContent({ albumId }: { albumId: string }) {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const { authHeader }: UserContextT = useContext(UserContext)

    const [albumData, setAlbumData]: [
        albumData: { albumMeta: AlbumData; media: WeblensMedia[] },
        setAlbumData: any,
    ] = useState(null)
    const mType = useMediaType()
    const [notFound, setNotFound] = useState(false)
    const nav = useNavigate()

    const fetchAlbum = useCallback(() => {
        if (!mType) {
            return
        }
        galleryDispatch({ type: 'add_loading', loading: 'album_media' })
        GetAlbumMedia(albumId, galleryState.includeRaw, authHeader)
            .then((m) => {
                galleryDispatch({ type: 'set_media', medias: m.media })
                setAlbumData(m)
            })
            .catch((r) => {
                if (r === 404) {
                    setNotFound(true)
                    return
                }
                notifications.show({
                    title: 'Failed to load album',
                    message: String(r),
                    color: 'red',
                })
            })
            .finally(() =>
                galleryDispatch({
                    type: 'remove_loading',
                    loading: 'album_media',
                })
            )
    }, [albumId, galleryState.includeRaw, mType])

    useEffect(() => {
        fetchAlbum()
    }, [fetchAlbum])

    const media = useMemo(() => {
        if (!albumData) {
            return []
        }

        const media = albumData.media
            .filter((v) => {
                if (galleryState.searchContent === '') {
                    return true
                }
                return v.MatchRecogTag(galleryState.searchContent)
            })
            .reverse()
        media.unshift()

        return media
    }, [albumData?.media, galleryState.searchContent])

    if (notFound) {
        return (
            <NotFound
                resourceType="Album"
                link="/albums"
                setNotFound={setNotFound}
            />
        )
    }

    if (!albumData) {
        return null
    }

    return (
        <div className="w-full">
            {media.length === 0 && <AlbumNoContent albumData={albumData} />}

            {media.length !== 0 && (
                <PhotoGallery
                    medias={media}
                    album={albumData.albumMeta}
                    fetchAlbum={fetchAlbum}
                />
            )}
        </div>
    )
}

function NewAlbum({ fetchAlbums }: { fetchAlbums: () => void }) {
    const [newAlbumName, setNewAlbumName] = useState(null)
    const { authHeader } = useContext(UserContext)

    console.log(newAlbumName)

    return (
        <div className="flex items-center h-14 w-40">
            {newAlbumName === null && (
                <WeblensButton
                    squareSize={40}
                    label="New Album"
                    centerContent
                    Left={IconLibraryPlus}
                    onClick={(e) => {
                        setNewAlbumName('')
                    }}
                />
            )}
            {newAlbumName !== null && (
                // <div className="flex flex-row w-10 items-center justify-center bg-dark-paper rounded p-2">
                <WeblensInput
                    value={newAlbumName}
                    height={40}
                    onComplete={(val) =>
                        createAlbum(val, authHeader)
                            .then(() => {
                                setNewAlbumName(null)
                                fetchAlbums()
                            })
                            .catch((r) => {
                                console.error(r)
                            })
                    }
                    closeInput={() => setNewAlbumName(null)}
                    buttonIcon={IconPlus}
                />
                // </div>
            )}
        </div>
    )
}

const AlbumsControls = ({ albumId, fetchAlbums }) => {
    const nav = useNavigate()
    const {
        galleryState: mediaState,
        galleryDispatch: mediaDispatch,
    }: GalleryContextT = useContext(GalleryContext)

    const click = useCallback(
        () =>
            mediaDispatch({
                type: 'set_raw_toggle',
                raw: !mediaState.includeRaw,
            }),
        [mediaDispatch, mediaState.includeRaw]
    )

    const setSize = useCallback(
        (s) => mediaDispatch({ type: 'set_image_size', size: s }),
        [mediaDispatch]
    )

    if (albumId === '') {
        return (
            <div className="p-2 ml-3">
                <NewAlbum fetchAlbums={fetchAlbums} />
            </div>
        )
    }

    return (
        <div className="flex flex-row w-full h-max items-center m-2 gap-4 ml-3">
            <div className="mr-5">
                <WeblensButton
                    squareSize={40}
                    centerContent
                    Left={IconArrowLeft}
                    onClick={() => nav('/albums')}
                />
            </div>

            <Divider orientation="vertical" className="mr-5 my-1" />

            <WeblensSlider
                value={mediaState.imageSize}
                width={200}
                height={35}
                min={150}
                max={500}
                callback={setSize}
            />

            <WeblensButton
                label="RAWs"
                allowRepeat
                toggleOn={mediaState.includeRaw}
                centerContent
                squareSize={35}
                onClick={click}
            />
        </div>
    )
}

function AlbumsHomeView({ fetchAlbums }: { fetchAlbums: () => void }) {
    const { galleryState } = useContext(GalleryContext)

    const albums = useMemo(() => {
        if (!galleryState) {
            return []
        }

        return Array.from(galleryState.albumsMap.values()).map((a) => {
            if (!a.CoverMedia) {
                a.CoverMedia = new WeblensMedia({ mediaId: a.Cover })
            }

            return a
        })
    }, [galleryState?.albumsMap])

    if (albums.length === 0) {
        return (
            <div className="flex justify-center items-center w-full h-80">
                <div className="flex flex-col items-center w-52 gap-1">
                    <p className="w-max text-xl"> You have no albums </p>
                    <NewAlbum fetchAlbums={fetchAlbums} />
                </div>
            </div>
        )
    } else {
        return (
            <div style={{ width: '100%', height: '100%', padding: 10 }}>
                <AlbumScroller albums={albums} />
            </div>
        )
    }
}

export function Albums({ selectedAlbum }: { selectedAlbum: string }) {
    const { authHeader }: UserContextT = useContext(UserContext)
    const { galleryDispatch } = useContext(GalleryContext)

    const fetchAlbums = useCallback(() => {
        galleryDispatch({ type: 'add_loading', loading: 'albums' })
        getAlbums(authHeader).then((val) => {
            galleryDispatch({ type: 'set_albums', albums: val })
            galleryDispatch({ type: 'remove_loading', loading: 'albums' })
        })
    }, [authHeader, galleryDispatch])

    useEffect(() => {
        fetchAlbums()
    }, [])

    return (
        <>
            <AlbumsControls albumId={selectedAlbum} fetchAlbums={fetchAlbums} />
            {selectedAlbum === '' && (
                <AlbumsHomeView fetchAlbums={fetchAlbums} />
            )}
            {selectedAlbum !== '' && <AlbumContent albumId={selectedAlbum} />}
        </>
    )
}
