import { Divider } from '@mantine/core'

import {
    IconArrowLeft,
    IconMinus,
    IconPlus,
    IconSearch,
    IconTrash,
    IconUserMinus,
    IconUsersPlus,
} from '@tabler/icons-react'

import { useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FixedSizeList } from 'react-window'
import { AutocompleteUsers } from '../api/ApiFetch'

import './albumStyle.scss'
import { useKeyDown, useResize } from '../components/hooks'
import WeblensButton from '../components/WeblensButton'
import WeblensInput from '../components/WeblensInput'

import { MediaImage } from '../Media/PhotoContainer'
import { GalleryContext } from '../Pages/Gallery/GalleryLogic'
import { AlbumData, UserInfoT } from '../types/Types'
import { DeleteAlbum, LeaveAlbum, ShareAlbum } from './AlbumQuery'
import { useSessionStore } from '../components/UserInfo'
import WeblensMedia from '../Media/Media'
import { useMediaStore } from '../Media/MediaStateControl'

function AlbumShareMenu({
    album,
    isOpen,
    closeShareMenu,
}: {
    album: AlbumData
    isOpen: boolean
    closeShareMenu: () => void
}) {
    const { user, auth } = useSessionStore()
    const [users, setUsers] = useState(album.sharedWith)
    const [userSearch, setUserSearch] = useState('')
    const [userSearchResults, setUserSearchResults] = useState([])

    useEffect(() => {
        if (userSearch === '') {
            setUserSearchResults([])
            return
        }
        AutocompleteUsers(userSearch, auth).then((r) => {
            r = r.filter(
                (u: UserInfoT) =>
                    u.username !== user.username && !users.includes(u.username)
            )
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
                            Icon={IconSearch}
                            value={userSearch}
                            valueCallback={setUserSearch}
                            onComplete={() => {}}
                        />
                    </div>

                    <div className="flex flex-col items-center justify-center h-[100px] w-full overflow-y-scroll p-1">
                        {userSearchResults.length === 0 && (
                            <p className="flex select-none text-white h-full items-center">
                                No Search Results
                            </p>
                        )}
                        {userSearchResults.map((u: UserInfoT) => {
                            console.log(u)
                            return (
                                <WeblensButton
                                    squareSize={40}
                                    label={u.username}
                                    Right={IconPlus}
                                    onClick={(e) => {
                                        e.stopPropagation()
                                        setUsers((p) => {
                                            const newP = [...p]
                                            newP.push(u.username)
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
                <div className="flex flex-col w-full items-center min-h-[150px]">
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
                                    Right={IconMinus}
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
                <div className="flex flex-row justify-around w-full h-max">
                    <div className="flex w-[48%] justify-center">
                        <WeblensButton
                            squareSize={40}
                            subtle
                            fillWidth
                            label={'Back'}
                            Left={IconArrowLeft}
                            onClick={() => {
                                closeShareMenu()
                            }}
                        />
                    </div>
                    <div className="flex w-[48%] justify-center">
                        <WeblensButton
                            squareSize={40}
                            fillWidth
                            label={'Share'}
                            Left={IconUsersPlus}
                            onClick={async () => {
                                return ShareAlbum(
                                    album.id,
                                    auth,
                                    users,
                                    album.sharedWith.filter(
                                        (u) => !users.includes(u)
                                    )
                                )
                                    .then((r) => {
                                        if (r.status !== 200) {
                                            return Promise.reject(
                                                'Failed to share album'
                                            )
                                        }
                                        setTimeout(() => closeShareMenu(), 1000)
                                        return true
                                    })
                                    .catch(() => false)
                            }}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export function MiniAlbumCover({
    album,
    disabled,
}: {
    album: AlbumData
    disabled?: boolean
}) {
    const mediaData = useMediaStore((state) => state.mediaMap.get(album.cover))

    return (
        <div
            key={album.id}
            className="album-selector animate-fade"
            data-included="true"
            data-has-media={Boolean(album.cover)}
            data-disabled={disabled}
        >
            <MediaImage media={mediaData} quality={'thumbnail'} />
            <p className="album-selector-title">{album.name}</p>
        </div>
    )
}

export function SingleAlbumCover({ album }: { album: AlbumData }) {
    const { galleryDispatch } = useContext(GalleryContext)
    const user = useSessionStore((state) => state.user)
    const auth = useSessionStore((state) => state.auth)
    const addMedias = useMediaStore((state) => state.addMedias)
    const nav = useNavigate()

    const [sharing, setSharing] = useState(false)
    const fontSize = Math.floor(Math.pow(0.975, album.name.length) * 40)
    const mediaData = useMediaStore((state) => state.mediaMap.get(album.cover))

    useEffect(() => {
        if (!mediaData) {
            const newM = new WeblensMedia({ contentId: album.cover })
            newM.LoadInfo().then(() => {
                addMedias([newM])
            })
        }
    }, [mediaData])

    useEffect(() => {
        galleryDispatch({ type: 'set_block_focus', block: sharing })
    }, [sharing])

    return (
        <div
            className="album-preview"
            onClick={() => {
                nav(album.id)
            }}
        >
            <div
                className="cover-box"
                data-faux-album={album.id === ''}
                data-no-cover={album.cover === ''}
                data-sharing={sharing.toString()}
            >
                <div
                    className="flex flex-col justify-end w-full h-full absolute"
                    style={{
                        pointerEvents: sharing ? 'none' : 'all',
                    }}
                >
                    <div className="album-title-wrapper">
                        <p
                            className="album-title-text truncate"
                            style={{ fontSize: fontSize }}
                        >
                            {album.name}
                        </p>

                        <div className="album-controls-wrapper">
                            <WeblensButton
                                subtle
                                squareSize={40}
                                Left={IconUsersPlus}
                                label={
                                    album.owner !== user.username
                                        ? `Shared by ${album.owner}`
                                        : ''
                                }
                                disabled={album.owner !== user.username}
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
                                    album.owner !== user.username
                                        ? IconUserMinus
                                        : IconTrash
                                }
                                onClick={(e) => {
                                    e.stopPropagation()

                                    let rq: Promise<Response>

                                    if (album.owner !== user.username) {
                                        rq = LeaveAlbum(album.id, auth).then()
                                    } else {
                                        rq = DeleteAlbum(album.id, auth)
                                    }

                                    rq.then((r) => {
                                        if (r.status === 200) {
                                            galleryDispatch({
                                                type: 'remove_album',
                                                albumId: album.id,
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
                    media={mediaData}
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
            thisData.push({ id: '', name: '' } as AlbumData)
        }

        return thisData
    }, [data, index])

    return (
        <div className="albums-row" style={style}>
            {thisData.map((a, i) => {
                if (a.id !== '') {
                    return <SingleAlbumCover key={a.id} album={a} />
                } else {
                    return (
                        <div key={`fake-album-${i}`} className="faux-album" />
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
        <div ref={setContainerRef} className="albums-container">
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
        </div>
    )
}
