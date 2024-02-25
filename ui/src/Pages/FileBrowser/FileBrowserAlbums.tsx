import { memo, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react"

import { Box, Button, Divider, Input, Loader, Space, Text, TextInput, Tooltip, TooltipFloating } from "@mantine/core"
import { IconExternalLink, IconPhoto, IconPlus, IconSearch, IconUsersGroup } from "@tabler/icons-react"
import { notifications } from "@mantine/notifications"

import { AddMediaToAlbum, CreateAlbum, GetAlbums } from "../../api/GalleryApi"
import { MediaImage } from "../../components/PhotoContainer"
import { AlbumData, getBlankMedia } from "../../types/Types"
import { ColumnBox, RowBox } from "./FilebrowserStyles"
import { userContext } from "../../Context"
import { VariableSizeList } from "react-window"
import { GetMediasByFolder } from "../../api/FileBrowserApi"

const useEnter = (cb) => {
    const onEnter = useCallback((e) => {
        if (e.key === "Enter") {
            cb()
        }
    }, [cb])

    useEffect(() => {
        document.addEventListener('keydown', onEnter)
        return () => {
            document.removeEventListener('keydown', onEnter)
        }
    }, [onEnter])
}

function NewAlbum({ refreshAlbums }: { refreshAlbums: (doLoading) => Promise<void> }) {
    const { authHeader } = useContext(userContext)

    const [newAlbumName, setNewAlbumName] = useState(null)
    const [loading, setLoading] = useState(false)

    const create = useCallback(() => { setLoading(true); CreateAlbum(newAlbumName, authHeader).then(() => { refreshAlbums(false).then(() => setNewAlbumName(null)); setLoading(false) }) }, [newAlbumName, authHeader, refreshAlbums])
    useEnter(create)

    return (
        <Box
            className="album-preview-row"
            style={{ height: '40px', margin: 0 }}
            onClick={() => { if (newAlbumName === null) { setNewAlbumName("") } }}
        >
            {loading && (
                <Box className="album-preview-loading" onClick={e => e.stopPropagation()} />
            )}
            {(newAlbumName == null && (
                <RowBox>
                    <IconPlus />
                    <Text size="16px" style={{ paddingLeft: 10 }}>New Album</Text>
                </RowBox>
            )) || (
                    <RowBox>
                        <Box className="media-placeholder" style={{ height: '50px', width: '50px' }}>
                            <IconPhoto />
                        </Box>
                        <TextInput variant="unstyled" size="16px" autoFocus onBlur={() => { if (!newAlbumName) { setNewAlbumName(null) } }} placeholder="Album name" value={newAlbumName} onChange={e => setNewAlbumName(e.target.value)} styles={{ input: { height: '30px' } }} style={{ lineHeight: 20, width: '100%' }} />
                    </RowBox>
                )}
        </Box>
    )
}

const SingleAlbum = memo(({ album, setMediaCallback, PartialApiCall, disabled = false }: { album: AlbumData, setMediaCallback: (mediaId: string, quality: "thumbnail" | "fullres", data: ArrayBuffer) => void, PartialApiCall: (albumId: string) => void, disabled?: boolean }) => {
    const { userInfo } = useContext(userContext)
    return (
        <Box className='album-preview-row' style={{ cursor: disabled ? 'default' : 'pointer', backgroundColor: disabled ? '#00000000' : '' }} onClick={e => { if (disabled) { e.stopPropagation(); return }; PartialApiCall(album.Id) }}>
            <MediaImage media={album.CoverMedia} setMediaCallback={setMediaCallback} quality='thumbnail' expectFailure={album.Cover === ""} containerStyle={{ borderRadius: '5px', overflow: 'hidden', width: '65px', height: '65px' }} disabled={disabled} />
            <RowBox style={{ width: '235px', justifyContent: 'space-evenly', flexGrow: 0 }}>
                <ColumnBox style={{ height: 'max-content', width: '50%', alignItems: 'flex-start', flexGrow: 1 }}>
                    <Box style={{ display: 'flex', flexGrow: 0, width: 'max-content', maxWidth: '100%', alignItems: 'center', paddingBottom: '10px' }}>
                        <Tooltip disabled={disabled} openDelay={200} label={album.Name}>
                            <Text c={disabled ? '#777777' : 'white'} size="16px" fw={disabled ? 450 : 550} truncate='end' styles={{ root: { width: '100%' } }}>{album.Name}</Text>
                        </Tooltip>
                        {album.Owner !== userInfo.username && (
                            <Tooltip label={`Shared by ${album.Owner}`}>
                                <IconUsersGroup color={disabled ? '#777777' : "white"} size={'20px'} style={{ marginLeft: 10 }} />
                            </Tooltip>
                        )}
                    </Box>
                    <RowBox>
                        <RowBox>
                            <IconPhoto color={disabled ? '#777777' : "white"} size={'15px'} />
                            <Text size='15px' c={disabled ? '#777777' : "white"} style={{ paddingLeft: 5 }}>{album.Medias.length}</Text>
                        </RowBox>

                    </RowBox>
                </ColumnBox>
                <RowBox style={{ position: 'absolute', width: 'max-content', alignItems: 'flex-end', justifyContent: 'flex-end', padding: 4, right: 0, cursor: 'pointer' }}>
                    <TooltipFloating position='right' label='Open Album'>
                        <IconExternalLink size={'15px'} onClick={(e) => { e.stopPropagation(); window.open(`/albums/${album.Id}`, '_blank') }} onMouseOver={(e) => { e.stopPropagation() }} />
                    </TooltipFloating>
                </RowBox>
            </RowBox>
        </Box>
    )
}, (prev, next) => {
    if (prev.disabled !== next.disabled) {
        return false
    }

    return false
})

const fetchAlbums = (doLoading, setLoading, setAlbums, authHeader) => {
    if (authHeader.Authorization === "") {
        return
    }
    if (doLoading) {
        setLoading(true)
    }
    return GetAlbums(authHeader).then(ret => {
        setAlbums((prev: AlbumData[]) => {
            if (!prev) {
                ret = ret.map(a => {
                    a.CoverMedia = getBlankMedia()
                    a.CoverMedia.fileHash = a.Cover

                    return a
                })
                return ret
            }
            const prevIds = prev.map(v => v.Id)
            for (const album of ret) {
                const i = prevIds.indexOf(album.Id)
                if (i !== -1) {
                    const mediaSave = prev[i].CoverMedia
                    prev[i] = album
                    prev[i].CoverMedia = mediaSave
                    prev[i].CoverMedia.fileHash = album.Cover
                } else {
                    if (!album.CoverMedia) {
                        album.CoverMedia = getBlankMedia()
                        album.CoverMedia.fileHash = album.Cover
                    }
                    prev.push(album)
                }
            }
            return [...prev]
        })
        setLoading(false)
    })
}

export const AlbumScoller = memo(({ candidates, authHeader }: {
    candidates: { media: string[], folders: string[] },
    authHeader
}) => {
    const [albums, setAlbums]: [albums: AlbumData[], setAlbums: any] = useState(null)
    const scrollboxRef = useRef(null)
    // This is for the state if we are waiting for the list of albums
    const [loading, setLoading] = useState(false)
    const [searchStr, setSearchStr] = useState("")

    // This is for tracking which album(s) are waiting
    // for results of adding media... naming is hard
    const [loadingAlbums, setLoadingAlbums] = useState([])

    const addMediaApiCall = useCallback((albumId) => {
        setLoadingAlbums(cur => [...cur, albumId])
        AddMediaToAlbum(albumId, candidates.media, candidates.folders, authHeader)
            .then((res) => {
                if (res.errors.length === 0) {
                    setLoadingAlbums(cur => cur.filter(v => v !== albumId))
                    fetchAlbums(false, setLoading, setAlbums, authHeader)
                    if (res.addedCount === 0) {
                        notifications.show({ message: `No new media to add to album`, color: 'orange' })
                    } else {
                        notifications.show({ message: `Added ${res.addedCount} medias to album`, color: 'green' })
                    }
                } else {
                    Promise.reject(res.errors)
                }
            })
            .catch((r) => { notifications.show({ title: "Could not add media to album", message: String(r), color: 'red' }) })
    }, [candidates, authHeader])

    const setMediaCallback = useCallback((mediaId: string, quality: "thumbnail" | "fullres", data: ArrayBuffer) => {
        setAlbums((prev: AlbumData[]) => {
            const mediaIds = prev.map(a => a.Cover)
            const i = mediaIds.indexOf(mediaId)
            if (i === -1) {
                return prev
            }

            if (!prev[i].CoverMedia) {
                prev[i].CoverMedia = getBlankMedia()
                prev[i].CoverMedia.fileHash = mediaId
            }
            prev[i].CoverMedia[quality] = data
            return [...prev]
        })
    }, [])

    useEffect(() => {
        console.log("HERE")
        fetchAlbums(true, setLoading, setAlbums, authHeader)
    }, [authHeader])

    const allMedias = useMemo(() => {
        const allMedias = []
        candidates.folders.forEach(f => GetMediasByFolder(f, authHeader).then(v => allMedias.push(...v.medias)))
        allMedias.push(...candidates.media)
        return allMedias
    }, [candidates.folders])

    const filteredAlbums = useMemo(() => albums?.filter(a => a.Name.toLowerCase().includes(searchStr)), [albums, searchStr])

    return (
        <ColumnBox style={{ maxHeight: 660, height: 'max-content', width: 320 }}>
            {(allMedias.length === 0 && (
                <Text size="20px" style={{ padding: 10 }}>No valid media selected</Text>
            )) || (
                    <Text size="20px" style={{ paddingBottom: 10 }}>Add {allMedias.length} item{allMedias.length === 1 ? '' : 's'} to albums</Text>
                )}
            <Input value={searchStr} onChange={e => setSearchStr(e.target.value.toLowerCase())} placeholder="Find an album" leftSection={<IconSearch color="#cccccc" size={'18px'} />} classNames={{ input: 'album-search' }} style={{ width: '100%', marginBottom: 8 }} />
            <NewAlbum refreshAlbums={(l: boolean) => fetchAlbums(l, setLoading, setAlbums, authHeader)} />
            <Divider my={10} w={'100%'} />
            {loading && (
                <Loader color="white" style={{ height: 'max-content', padding: 20 }} />
            )}
            <VariableSizeList
                className='no-scrollbars'
                ref={scrollboxRef}
                itemCount={filteredAlbums?.length ? filteredAlbums.length : 0}
                itemSize={i => 75}
                itemData={filteredAlbums}
                height={75 * filteredAlbums?.length < 600 ? 75 * filteredAlbums.length : 0}
                width={'100%'}
                itemKey={(index: number, data: AlbumData[]) => data[index]?.Id}
            >
                {({ data, index, style }) => AlbumRowWrap(data, index, style, allMedias, loadingAlbums, setMediaCallback, addMediaApiCall, authHeader)}
            </VariableSizeList>
        </ColumnBox>
    )
}, (prev, next) => {
    return true
})

const AlbumRowWrap = (data: AlbumData[], index, style, allMedias: string[], loadingAlbums, setMediaCallback, create, authHeader) => {
    const disabled = allMedias.every(m => data[index].Medias.includes(m))
    return (
        <Box style={style}>
            <SingleAlbum setMediaCallback={setMediaCallback} album={data[index]} PartialApiCall={create} disabled={disabled || loadingAlbums.includes(data[index].Id)} />
        </Box>
    )

}