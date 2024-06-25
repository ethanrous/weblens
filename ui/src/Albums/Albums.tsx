import { Divider, Space, Text } from '@mantine/core'
import {
    IconArrowLeft,
    IconFolder,
    IconInputX,
    IconLibraryPlus,
    IconPlus,
    IconSearch,
} from '@tabler/icons-react'
import { useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMediaType } from '../components/hooks'
import NotFound from '../components/NotFound'
import WeblensButton from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'
import WeblensSlider from '../components/WeblensSlider'
import { UserContext } from '../Context'
import { PhotoGallery } from '../Media/MediaDisplay'
import { GalleryContext, GalleryContextT } from '../Pages/Gallery/GalleryLogic'
import WeblensMedia from '../Media/Media'

import { AlbumData, UserContextT } from '../types/Types'
import { AlbumScroller } from './AlbumDisplay'
import { createAlbum, getAlbumMedia, getAlbums } from './AlbumQuery'

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
                {albumData.albumMeta.name}
            </Text>
            <div className="flex flex-col pt-40 w-max items-center">
                {albumData.albumMeta.medias.length !== 0 && (
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
                {albumData.albumMeta.medias.length === 0 && (
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
        getAlbumMedia(albumId, galleryState.includeRaw, authHeader)
            .then((m) => {
                setAlbumData(m)
            })
            .catch((r) => {
                if (r === 404) {
                    setNotFound(true)
                    return
                }
                console.error(r)
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
        if (albumData.media) {
            const media = albumData.media
                ?.filter((v) => {
                    if (galleryState.searchContent === '') {
                        return true
                    }
                    return v.MatchRecogTag(galleryState.searchContent)
                })
                .reverse()
            media?.unshift()
            return media
        }

        return []
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
                    autoFocus
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
    const { galleryState, galleryDispatch }: GalleryContextT =
        useContext(GalleryContext)

    const click = useCallback(
        () =>
            galleryDispatch({
                type: 'set_raw_toggle',
                raw: !galleryState.includeRaw,
            }),
        [galleryDispatch, galleryState.includeRaw]
    )

    const setSize = useCallback(
        (s) => galleryDispatch({ type: 'set_image_size', size: s }),
        [galleryDispatch]
    )

    if (albumId === '') {
        return (
            <div className="flex items-center justify-between p-2 ml-3">
                <NewAlbum fetchAlbums={fetchAlbums} />
                <div className="flex items-center h-[40px] w-60 pr-6">
                    <WeblensInput
                        value={galleryState.searchContent}
                        Icon={IconSearch}
                        stealFocus={!galleryState.blockSearchFocus}
                        height={40}
                        valueCallback={(v) =>
                            galleryDispatch({ type: 'set_search', search: v })
                        }
                        onComplete={() => {}}
                    />
                </div>
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
                value={galleryState.imageSize}
                width={200}
                height={35}
                min={150}
                max={500}
                callback={setSize}
            />

            <WeblensButton
                label="RAWs"
                allowRepeat
                toggleOn={galleryState.includeRaw}
                centerContent
                squareSize={35}
                onClick={click}
            />
        </div>
    )
}

function AlbumsHomeView({ fetchAlbums }: { fetchAlbums: () => void }) {
    const { galleryState, galleryDispatch } = useContext(GalleryContext)

    const albums = useMemo(() => {
        if (!galleryState) {
            return []
        }

        return Array.from(galleryState.albumsMap.values()).filter((a) =>
            a.name
                .toLowerCase()
                .includes(galleryState.searchContent.toLowerCase())
        )
        // .map((a) => {
        // if (!a.CoverMedia) {
        //     a.CoverMedia = new WeblensMedia({ contentId: a.Cover })
        // }
        //
        // return a
        // })
    }, [galleryState?.albumsMap, galleryState.searchContent])

    if (albums.length === 0 && galleryState.searchContent === '') {
        return (
            <div className="flex justify-center items-center w-full h-80 select-none">
                <div className="flex flex-col items-center w-52 gap-1">
                    <p className="w-max text-xl"> You have no albums </p>
                    <NewAlbum fetchAlbums={fetchAlbums} />
                </div>
            </div>
        )
    } else if (albums.length === 0 && galleryState.searchContent !== '') {
        return (
            <div className="flex flex-col justify-center items-center w-full h-80 select-none gap-2">
                <div className="flex items-center w-max gap-1">
                    <p className="w-max text-xl text-nowrap">
                        No albums found with
                    </p>
                    <div className="flex flex-row items-center bg-dark-paper rounded pl-1 pr-1 pt-[1px] pb-[1px] gap-1">
                        <IconSearch size={18} />
                        <p className="w-max text-xl text-nowrap">
                            {galleryState.searchContent}
                        </p>
                    </div>
                    <p className="w-max text-xl text-nowrap">in their name</p>
                </div>
                <div className="flex items-center">
                    <WeblensButton
                        label={'Clear Search'}
                        Left={IconInputX}
                        onClick={() => {
                            galleryDispatch({ type: 'set_search', search: '' })
                        }}
                    />
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
        getAlbums(true, authHeader).then((val) => {
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
