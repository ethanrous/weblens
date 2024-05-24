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

import { CreateAlbum, GetAlbumMedia, GetAlbums } from '../../api/GalleryApi'
import { AlbumData, AuthHeaderT, UserContextT } from '../../types/Types'
import WeblensMedia from '../../classes/Media'
import { UserContext } from '../../Context'
import { PhotoGallery } from '../../components/MediaDisplay'
import NotFound from '../../components/NotFound'

import { GalleryContextT } from './Gallery'
import { useMediaType } from '../../components/hooks'
import { AlbumScroller } from './AlbumDisplay'
import { GalleryContext } from './Gallery'
import { WeblensButton } from '../../components/WeblensButton'
import WeblensInput from '../../components/WeblensInput'
import WeblensSlider from '../../components/WeblensSlider'

// function ShareBox({
//     open,
//     setOpen,
//     albumId,
//     sharedWith,
//     fetchAlbums,
// }: {
//     open: boolean;
//     setOpen;
//     albumId;
//     sharedWith;
//     fetchAlbums;
// }) {
//     const { galleryState } = useContext(GalleryContext);
//     const { authHeader }: UserContextT = useContext(UserContext);
//     const [value, setValue] = useState(sharedWith);

//     useEffect(() => {
//         setValue(sharedWith);
//     }, [sharedWith]);

//     return (
//         <Popover
//             opened={open}
//             onClose={() => setOpen(false)}
//             closeOnClickOutside
//         >
//             <Popover.Target>
//                 <Box
//                     style={{
//                         position: "fixed",
//                         top: galleryState.menuPos.y,
//                         left: galleryState.menuPos.x,
//                     }}
//                 />
//             </Popover.Target>
//             <Popover.Dropdown>
//                 {/* <ShareInput valueSetCallback={setValue} initValues={sharedWith} /> */}
//                 <Space h={10} />
//                 <Button
//                     fullWidth
//                     disabled={
//                         JSON.stringify(value) === JSON.stringify(sharedWith)
//                     }
//                     color="#4444ff"
//                     onClick={() => {
//                         ShareAlbum(
//                             albumId,
//                             authHeader,
//                             value.filter((v) => !sharedWith.includes(v)),
//                             sharedWith.filter((v) => !value.includes(v))
//                         ).then(() => fetchAlbums());
//                         setOpen(false);
//                     }}
//                 >
//                     Update
//                 </Button>
//             </Popover.Dropdown>
//         </Popover>
//     );
// }

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

    if (media.length === 0) {
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
                        height={40}
                        centerContent
                        label="FileBrowser"
                        Left={<IconFolder />}
                        // postScript="Upload new media"
                        width={400}
                        onClick={() => nav('/files/home')}
                    />
                </div>
            </div>
        )
    }

    return (
        <div className="w-full">
            <Space h={10} />
            <AlbumsControls albumId={albumId} />
            <PhotoGallery
                medias={media}
                album={albumData.albumMeta}
                fetchAlbum={fetchAlbum}
            />
        </div>
    )
}

const AlbumsControls = ({ albumId }) => {
    const [newAlbumModal, setNewAlbumModal] = useState(false)
    const [newAlbumName, setNewAlbumName] = useState('')
    const {
        galleryState: mediaState,
        galleryDispatch: mediaDispatch,
    }: GalleryContextT = useContext(GalleryContext)
    const { authHeader }: UserContextT = useContext(UserContext)

    const click = useCallback(
        () =>
            mediaDispatch({
                type: 'set_raw_toggle',
                raw: !mediaState.includeRaw,
            }),
        [mediaDispatch, mediaState.includeRaw]
    )

    // useEffect(() => {}, [mediaDispatch, mediaState.includeRaw, click]);

    const setSize = useCallback(
        (s) => mediaDispatch({ type: 'set_image_size', size: s }),
        [mediaDispatch]
    )

    return (
        <div className="flex flex-row w-full h-max items-center">
            {albumId !== '' && (
                <div className="mr-10 ml-3">
                    <WeblensButton
                        height={40}
                        width={40}
                        centerContent
                        Left={<IconArrowLeft width={24} height={24} />}
                    />
                </div>
            )}
            <WeblensSlider
                value={mediaState.imageSize}
                width={200}
                height={35}
                min={150}
                max={500}
                callback={setSize}
            />

            <Space w={20} />
            <WeblensButton
                label="RAWs"
                allowRepeat
                toggleOn={mediaState.includeRaw}
                centerContent
                height={35}
                width={80}
                onClick={click}
            />
        </div>
    )
}

function AlbumsHomeView({}) {
    const { authHeader }: UserContextT = useContext(UserContext)
    const { galleryState, galleryDispatch } = useContext(GalleryContext)
    const [newAlbumName, setNewAlbumName] = useState(null)

    const fetchAlbums = useCallback(() => {
        galleryDispatch({ type: 'add_loading', loading: 'albums' })
        GetAlbums(authHeader).then((val) => {
            galleryDispatch({ type: 'set_albums', albums: val })
            galleryDispatch({ type: 'remove_loading', loading: 'albums' })
        })
    }, [authHeader, galleryDispatch])

    useEffect(() => {
        fetchAlbums()
    }, [])

    const albums = Array.from(galleryState.albumsMap.values()).map((a) => {
        if (!a.CoverMedia) {
            a.CoverMedia = new WeblensMedia({ mediaId: a.Cover })
        }

        return a
    })

    if (albums.length === 0) {
        return (
            <div className="flex justify-center items-center w-full h-80">
                <div className="flex flex-col items-center w-52 gap-1">
                    <p className="w-max text-xl"> You have no albums </p>
                    <div className="h-[50px]">
                        {newAlbumName === null && (
                            <WeblensButton
                                height={40}
                                label="New Album"
                                centerContent
                                Left={<IconLibraryPlus />}
                                onClick={(e) => {
                                    setNewAlbumName('')
                                }}
                            />
                        )}
                        {newAlbumName !== null && (
                            <div className="flex flex-row items-center justify-center bg-dark-paper rounded p-2">
                                <WeblensInput
                                    value={newAlbumName}
                                    onComplete={(val) =>
                                        CreateAlbum(val, authHeader)
                                    }
                                    closeInput={() => setNewAlbumName(null)}
                                />
                                {/*<input*/}
                                {/*    autoFocus*/}
                                {/*    value={newAlbumName}*/}
                                {/*    onChange={(event) =>*/}
                                {/*        setNewAlbumName(event.target.value)*/}
                                {/*    }*/}
                                {/*    onBlur={(event) => setNewAlbumName(null)}*/}
                                {/*    className="weblens-input h-[40px]"*/}
                                {/*/>*/}
                                <WeblensButton
                                    height={40}
                                    Left={<IconPlus />}
                                    centerContent
                                />
                            </div>
                        )}
                    </div>
                </div>
            </div>
        )
    } else {
        return (
            <div style={{ width: '100%', height: '100%', padding: 10 }}>
                <AlbumsControls albumId={''} />
                {/* <AlbumCoverMenu fetchAlbums={fetchAlbums} /> */}
                <AlbumScroller albums={albums} />
            </div>
        )
    }
}

export function Albums({ selectedAlbum }: { selectedAlbum: string }) {
    if (selectedAlbum === '') {
        return <AlbumsHomeView />
    } else {
        return <AlbumContent albumId={selectedAlbum} />
    }
}
