import { Box, Divider, Text } from '@mantine/core'

import { useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FixedSizeList } from 'react-window'
import { useKeyDown, useResize } from '../components/hooks'

import {
    IconArrowLeft,
    IconMinus,
    IconPlus,
    IconSearch,
    IconTrash,
    IconUserMinus,
    IconUsersPlus,
} from '@tabler/icons-react'

import { MediaImage } from '../Media/PhotoContainer'
import { AlbumData, AuthHeaderT, UserContextT } from '../types/Types'
import WeblensMedia from '../Media/Media'
import { UserContext } from '../Context'
import { GalleryContext } from '../Pages/Gallery/Gallery'
import { WeblensButton } from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'
import { AutocompleteUsers } from '../api/ApiFetch'

import './albumStyle.scss'
import {
    DeleteAlbum,
    GetAlbumMedia,
    LeaveAlbum,
    SetAlbumCover,
    ShareAlbum,
} from './AlbumQuery'

function AlbumShareMenu({
    album,
    isOpen,
    closeShareMenu,
}: {
    album: AlbumData
    isOpen: boolean
    closeShareMenu: () => void
}) {
    const { usr, authHeader } = useContext(UserContext)
    const [users, setUsers] = useState(album.SharedWith)
    const [userSearch, setUserSearch] = useState('')
    const [userSearchResults, setUserSearchResults] = useState([])

    useEffect(() => {
        if (userSearch === '') {
            setUserSearchResults([])
            return
        }
        AutocompleteUsers(userSearch, authHeader).then((r) => {
            r = r.filter((u) => u !== usr.username && !users.includes(u))
            setUserSearchResults(r)
        })
    }, [userSearch, users])

    useKeyDown('Escape', () => closeShareMenu())

    if (!album) {
        return null
    }

    return (
        <div
            className="album-share-menu"
            onClick={(e) => {
                e.stopPropagation()
            }}
        >
            <div className="flex flex-col h-full w-full items-center justify-between bg-bottom-grey rounded p-1 cursor-default">
                <div className="flex flex-col w-full max-h-40 overflow-y-scroll rounded items-center p-2">
                    <div className="flex flex-row items-center h-10 w-48">
                        <WeblensInput
                            placeholder={'Search Users'}
                            icon={<IconSearch size={16} />}
                            value={userSearch}
                            valueCallback={setUserSearch}
                            onComplete={() => {}}
                        />
                    </div>

                    <div className="flex flex-col items-center justify-center h-[150px] w-full overflow-y-scroll p-1">
                        {userSearchResults.length === 0 && (
                            <p className="flex select-none text-white h-full items-center">
                                No Search Results
                            </p>
                        )}
                        {userSearchResults.map((u) => {
                            return (
                                <WeblensButton
                                    squareSize={40}
                                    label={u}
                                    Right={<IconPlus />}
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        setUsers((p) => {
                                            const newP = [...p]
                                            newP.push(u)
                                            return newP
                                        })
                                        setUserSearchResults((p) => {
                                            const newP = [...p]
                                            newP.splice(newP.indexOf(u), 1)
                                            return newP
                                        })
                                    }}
                                />
                            )
                        })}
                    </div>
                </div>
                <div className="flex flex-col w-full items-center">
                    <Divider orientation={'horizontal'} w={'95%'} />
                    <p className="m-2 text-white">Shared With</p>
                    <div
                        className="flex flex-col w-full max-h-20 overflow-y-scroll items-center p-1 cursor-default"
                        onClick={(e) => e.stopPropagation()}
                    >
                        {users.length === 0 && (
                            <p className="select-none text-white">Not Shared</p>
                        )}
                        {users.map((u) => {
                            return (
                                <WeblensButton
                                    key={u}
                                    subtle
                                    squareSize={40}
                                    label={u}
                                    Right={<IconMinus />}
                                    onClick={() => {
                                        setUsers((p) => {
                                            const newP = [...p]
                                            newP.splice(newP.indexOf(u), 1)
                                            return newP
                                        })
                                    }}
                                />
                            )
                        })}
                    </div>
                </div>
                <div className="flex flex-row w-full h-max">
                    <WeblensButton
                        squareSize={40}
                        subtle
                        label={'Cancel'}
                        Left={IconArrowLeft}
                        onClick={() => {
                            closeShareMenu()
                        }}
                    />
                    <WeblensButton
                        squareSize={40}
                        label={'Update'}
                        Left={IconUsersPlus}
                        onClick={async () => {
                            return ShareAlbum(
                                album.Id,
                                authHeader,
                                users,
                                album.SharedWith.filter(
                                    (u) => !users.includes(u)
                                )
                            )
                                .then(() => {
                                    setTimeout(() => closeShareMenu(), 1000)
                                    return true
                                })
                                .catch(() => false)
                        }}
                    />
                </div>
            </div>
        </div>
    )
}

function AlbumContentPreview({
    albumMedias,
    setCoverM,
    selected,
}: {
    albumMedias: WeblensMedia[]
    setCoverM
    selected: string
}) {
    if (albumMedias === null) {
        return null
    }

    return (
        <div className="content-preview-wrapper">
            <div className="overflow-x-scroll">
                <div className="flex flex-row h-max w-max gap-1">
                    {albumMedias.map((m) => {
                        return (
                            <div
                                key={m.Id()}
                                className="content-preview-item"
                                data-selected={selected === m.Id()}
                                onClick={(e) => {
                                    e.stopPropagation()
                                    if (selected !== m.Id()) {
                                        setCoverM(m)
                                    }
                                }}
                            >
                                <MediaImage
                                    key={m.Id()}
                                    media={m}
                                    quality={'thumbnail'}
                                    disabled={selected === m.Id()}
                                />
                            </div>
                        )
                    })}
                </div>
            </div>
        </div>
    )
}

export function MiniAlbumCover({
    album,
    disabled,
    authHeader,
}: {
    album: AlbumData
    disabled?: boolean
    authHeader?: AuthHeaderT
}) {
    if (!album.CoverMedia && album.Cover && authHeader) {
        album.CoverMedia = new WeblensMedia({ mediaId: album.Cover })
    }

    return (
        <div
            key={album.Id}
            className="album-selector animate-fade"
            data-included="true"
            data-has-media={Boolean(album.CoverMedia)}
            data-disabled={disabled}
        >
            <MediaImage media={album.CoverMedia} quality={'thumbnail'} />
            <p className="album-selector-title">{album.Name}</p>
        </div>
    )
}

export function SingleAlbumCover({ album }: { album: AlbumData }) {
    const { galleryDispatch } = useContext(GalleryContext)
    const { usr } = useContext(UserContext)
    const nav = useNavigate()
    const { authHeader }: UserContextT = useContext(UserContext)
    const [coverM, setCoverM] = useState(album.CoverMedia)
    const [previewMedia, setPreviewMedia] = useState<WeblensMedia[]>(null)
    const [sharing, setSharing] = useState(false)
    const fontSize = Math.floor(Math.pow(0.975, album.Name.length) * 40)

    useEffect(() => {
        galleryDispatch({ type: 'set_block_focus', block: sharing })
    }, [sharing])

    return (
        <div
            className="album-preview"
            onMouseOver={() => {
                if (previewMedia === null && album.Medias.length !== 0) {
                    GetAlbumMedia(album.Id, false, authHeader).then((val) =>
                        setPreviewMedia(val.media)
                    )
                }
            }}
            onClick={() => {
                nav(album.Id)
            }}
        >
            <div
                className="cover-box"
                data-faux-album={album.Id === ''}
                data-no-cover={coverM?.Id() === ''}
                data-sharing={sharing.toString()}
            >
                <div
                    className="flex flex-col justify-end w-full h-full absolute"
                    style={{
                        pointerEvents: sharing ? 'none' : 'all',
                    }}
                >
                    <AlbumContentPreview
                        albumMedias={previewMedia}
                        setCoverM={(media: WeblensMedia) => {
                            SetAlbumCover(
                                album.Id,
                                media.Id(),
                                authHeader
                            ).then(() => setCoverM(media))
                        }}
                        selected={coverM?.Id()}
                    />
                    <div className="album-title-wrapper">
                        <Text
                            truncate="end"
                            className="album-title-text"
                            size={`${fontSize}px`}
                        >
                            {album.Name}
                        </Text>

                        <div className="album-controls-wrapper">
                            <WeblensButton
                                subtle
                                squareSize={40}
                                Left={IconUsersPlus}
                                label={
                                    album.Owner !== usr.username
                                        ? `Shared by ${album.Owner}`
                                        : ''
                                }
                                disabled={album.Owner !== usr.username}
                                onClick={(e) => {
                                    e.stopPropagation()
                                    setSharing(true)
                                }}
                            />
                            <WeblensButton
                                subtle
                                danger
                                squareSize={40}
                                Left={
                                    album.Owner !== usr.username
                                        ? IconUserMinus
                                        : IconTrash
                                }
                                onClick={(e) => {
                                    e.stopPropagation()

                                    let rq: Promise<Response>

                                    if (album.Owner !== usr.username) {
                                        rq = LeaveAlbum(
                                            album.Id,
                                            authHeader
                                        ).then()
                                    } else {
                                        rq = DeleteAlbum(album.Id, authHeader)
                                    }

                                    rq.then((r) => {
                                        if (r.status === 200) {
                                            galleryDispatch({
                                                type: 'remove_album',
                                                albumId: album.Id,
                                            })
                                        } else {
                                            return false
                                        }
                                    })
                                }}
                            />
                        </div>
                    </div>
                </div>
                <MediaImage
                    media={coverM}
                    quality={'fullres'}
                    imgStyle={{ zIndex: -1 }}
                    containerClass="cover-image"
                />
                <AlbumShareMenu
                    album={album}
                    isOpen={sharing}
                    closeShareMenu={() => setSharing(false)}
                />
            </div>
        </div>
    )
}

function AlbumWrapper({
    data,
    index,
    style,
}: {
    data: { albums: AlbumData[]; colCount: number }
    index: number
    style
}) {
    const thisData = useMemo(() => {
        const thisData = data.albums.slice(
            index * data.colCount,
            index * data.colCount + data.colCount
        )

        while (thisData.length < data.colCount) {
            thisData.push({ Id: '', Name: '' } as AlbumData)
        }

        return thisData
    }, [data, index])

    return (
        <div className="albums-row" style={style}>
            {thisData.map((a, i) => {
                if (a.Id !== '') {
                    return <SingleAlbumCover key={a.Id} album={a} />
                } else {
                    return (
                        <Box key={`fake-album-${i}`} className="faux-album" />
                    )
                }
            })}
        </div>
    )
}

const ALBUM_WIDTH = 450

export function AlbumScroller({ albums }: { albums: AlbumData[] }) {
    const [containerRef, setContainerRef] = useState(null)
    const containerSize = useResize(containerRef)

    const colCount = Math.floor(containerSize.width / ALBUM_WIDTH)

    return (
        <Box ref={setContainerRef} className="albums-container">
            <FixedSizeList
                className="no-scrollbar"
                height={
                    containerSize.height >= 21 ? containerSize.height - 21 : 0
                }
                width={containerSize.width}
                itemSize={480}
                itemCount={
                    colCount !== 0 ? Math.ceil(albums.length / colCount) : 0
                }
                itemData={{
                    albums: albums,
                    colCount: colCount,
                }}
                overscanRowCount={5}
            >
                {AlbumWrapper}
            </FixedSizeList>
        </Box>
    )
}
