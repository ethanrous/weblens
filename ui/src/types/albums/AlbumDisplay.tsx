// import { Divider } from '@mantine/core'
// import {
//     IconArrowLeft,
//     IconMinus,
//     IconPlus,
//     IconSearch,
//     IconTrash,
//     IconUserMinus,
//     IconUsersPlus,
// } from '@tabler/icons-react'
// import { useQuery } from '@tanstack/react-query'
// import AlbumsApi from '@weblens/api/AlbumsApi'
// import SharesApi from '@weblens/api/SharesApi'
// import UsersApi from '@weblens/api/UserApi'
// import { AlbumInfo, UserInfo } from '@weblens/api/swag'
// import WeblensButton from '@weblens/lib/WeblensButton'
// import WeblensInput from '@weblens/lib/WeblensInput'
// import '@weblens/types/albums/albumStyle.scss'
// import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
// import { useMediaStore } from '@weblens/types/media/MediaStateControl'
// import { MediaImage } from '@weblens/types/media/PhotoContainer'
// import { useSessionStore } from 'components/UserInfo'
// import { useKeyDown, useResize } from '@weblens/lib/hooks'
// import { useContext, useEffect, useMemo, useState } from 'react'
// import { useNavigate } from 'react-router-dom'
// import { FixedSizeList } from 'react-window'
// import { ErrorHandler } from 'types/Types'
//
// function AlbumShareMenu({
//     album,
//     closeShareMenu,
// }: {
//     album: AlbumInfo
//     closeShareMenu: () => void
// }) {
//     const [users, setUsers] = useState<string[]>([])
//     const [userSearch, setUserSearch] = useState('')
//     const [lastSearch, setLastSearch] = useState('')
//
//     const { data: userSearchResults } = useQuery<UserInfo[]>({
//         queryKey: ['albumUsersSearch', userSearch],
//         initialData: [],
//         queryFn: async (ctx) => {
//             if (ctx.queryKey[1] === lastSearch) {
//                 const ret: UserInfo[] = userSearchResults.filter(
//                     (u) => !users.includes(u.username)
//                 )
//                 return ret
//             }
//             if (userSearch.length < 3) {
//                 return [] as UserInfo[]
//             }
//
//             const res = await UsersApi.searchUsers(userSearch, {
//                 signal: ctx.signal,
//             }).then((r) => {
//                 const res = r.data.filter((u) => !users.includes(u.username))
//                 setLastSearch(userSearch)
//                 return res
//             })
//             return res
//         },
//     })
//
//     useKeyDown('Escape', () => closeShareMenu())
//
//     if (!album) {
//         return null
//     }
//
//     return (
//         <div
//             className="album-share-menu"
//             onClick={(e) => {
//                 e.stopPropagation()
//             }}
//         >
//             <div className="flex flex-col h-full w-full items-center justify-between bg-bottom-grey rounded-sm p-1 cursor-default">
//                 <div className="flex flex-col w-full max-h-40 overflow-y-scroll rounded-sm items-center p-2">
//                     <div className="flex flex-row items-center h-10 w-48">
//                         <WeblensInput
//                             placeholder={'Search Users'}
//                             Icon={IconSearch}
//                             value={userSearch}
//                             valueCallback={setUserSearch}
//                             onComplete={async () => {}}
//                         />
//                     </div>
//
//                     <div className="flex flex-col items-center justify-center h-[100px] w-full overflow-y-scroll p-1">
//                         {userSearchResults.length === 0 && (
//                             <p className="flex select-none text-white h-full items-center">
//                                 No Search Results
//                             </p>
//                         )}
//                         {userSearchResults.map((u: UserInfo) => {
//                             return (
//                                 <WeblensButton
//                                     squareSize={40}
//                                     label={u.username}
//                                     Right={IconPlus}
//                                     onClick={(e) => {
//                                         e.stopPropagation()
//                                         setUsers((p) => {
//                                             const newP = [...p]
//                                             newP.push(u.username)
//                                             return newP
//                                         })
//                                     }}
//                                 />
//                             )
//                         })}
//                     </div>
//                 </div>
//                 <div className="flex flex-col w-full items-center min-h-[150px]">
//                     <Divider orientation={'horizontal'} w={'95%'} />
//                     <p className="m-2 text-white">Shared With</p>
//                     <div
//                         className="flex flex-col w-full max-h-20 overflow-y-scroll items-center p-1 cursor-default"
//                         onClick={(e) => e.stopPropagation()}
//                     >
//                         {!users ||
//                             (users.length === 0 && (
//                                 <p className="select-none text-white">
//                                     Not Shared
//                                 </p>
//                             ))}
//                         {users.map((u) => {
//                             return (
//                                 <WeblensButton
//                                     key={u}
//                                     subtle
//                                     squareSize={40}
//                                     label={u}
//                                     Right={IconMinus}
//                                     onClick={() => {
//                                         setUsers((p) => {
//                                             const newP = [...p]
//                                             newP.splice(newP.indexOf(u), 1)
//                                             return newP
//                                         })
//                                     }}
//                                 />
//                             )
//                         })}
//                     </div>
//                 </div>
//                 <div className="flex flex-row justify-around w-full h-max">
//                     <div className="flex w-[48%] justify-center">
//                         <WeblensButton
//                             squareSize={40}
//                             subtle
//                             fillWidth
//                             label={'Back'}
//                             Left={IconArrowLeft}
//                             onClick={() => {
//                                 closeShareMenu()
//                             }}
//                         />
//                     </div>
//                     <div className="flex w-[48%] justify-center">
//                         <WeblensButton
//                             squareSize={40}
//                             fillWidth
//                             label={'share'}
//                             Left={IconUsersPlus}
//                             onClick={async () => {
//                                 return SharesApi.createAlbumShare({
//                                     albumId: album.id,
//                                     public: false,
//                                     users: users,
//                                 }).then(() => {
//                                     setTimeout(() => closeShareMenu(), 1000)
//                                 })
//                             }}
//                         />
//                     </div>
//                 </div>
//             </div>
//         </div>
//     )
// }
//
// export function MiniAlbumCover({
//     album,
//     disabled,
// }: {
//     album: AlbumInfo
//     disabled?: boolean
// }) {
//     const mediaData = useMediaStore((state) => state.mediaMap.get(album.cover))
//
//     return (
//         <div
//             key={album.id}
//             className="album-selector animate-fade"
//             data-included="true"
//             data-has-media={Boolean(album.cover)}
//             data-disabled={disabled}
//         >
//             <MediaImage media={mediaData} quality={PhotoQuality.LowRes} />
//             <p className="album-selector-title">{album.name}</p>
//         </div>
//     )
// }
//
// export function SingleAlbumCover({ album }: { album: AlbumInfo }) {
//     const { galleryDispatch } = useContext(GalleryContext)
//     const user = useSessionStore((state) => state.user)
//     const addMedias = useMediaStore((state) => state.addMedias)
//     const nav = useNavigate()
//
//     const [sharing, setSharing] = useState(false)
//     const fontSize = Math.floor(Math.pow(0.975, album.name.length) * 40)
//     const mediaData = useMediaStore((state) => state.mediaMap.get(album.cover))
//
//     useEffect(() => {
//         if (!mediaData && album.cover) {
//             const newM = new WeblensMedia({ contentId: album.cover })
//             newM.LoadInfo()
//                 .then(() => {
//                     addMedias([newM])
//                 })
//                 .catch((err) => console.error('Failed to load media info', err))
//         }
//     }, [mediaData])
//
//     useEffect(() => {
//         galleryDispatch({ type: 'set_block_focus', block: sharing })
//     }, [sharing])
//
//     return (
//         <div
//             className="album-preview"
//             onClick={() => {
//                 nav(album.id)
//             }}
//         >
//             <div
//                 className="cover-box"
//                 data-faux-album={album.id === ''}
//                 data-no-cover={album.cover === ''}
//                 data-sharing={sharing.toString()}
//             >
//                 <div
//                     className="flex flex-col justify-end w-full h-full absolute"
//                     style={{
//                         pointerEvents: sharing ? 'none' : 'all',
//                     }}
//                 >
//                     <div className="album-title-wrapper">
//                         <p
//                             className="album-title-text truncate"
//                             style={{ fontSize: fontSize }}
//                         >
//                             {album.name}
//                         </p>
//
//                         <div className="album-controls-wrapper">
//                             <WeblensButton
//                                 subtle
//                                 squareSize={40}
//                                 Left={IconUsersPlus}
//                                 label={
//                                     album.owner !== user.username
//                                         ? `Shared by ${album.owner}`
//                                         : ''
//                                 }
//                                 disabled={album.owner !== user.username}
//                                 onClick={(e) => {
//                                     e.stopPropagation()
//                                     setSharing(true)
//                                 }}
//                             />
//                             <WeblensButton
//                                 subtle
//                                 danger
//                                 squareSize={40}
//                                 Left={
//                                     album.owner !== user.username
//                                         ? IconUserMinus
//                                         : IconTrash
//                                 }
//                                 onClick={(e) => {
//                                     e.stopPropagation()
//
//                                     AlbumsApi.deleteOrLeaveAlbum(album.id)
//                                         .then((r) => {
//                                             if (r.status === 200) {
//                                                 galleryDispatch({
//                                                     type: 'remove_album',
//                                                     albumId: album.id,
//                                                 })
//                                             } else {
//                                                 return false
//                                             }
//                                         })
//                                         .catch(ErrorHandler)
//                                 }}
//                             />
//                         </div>
//                     </div>
//                 </div>
//                 <MediaImage
//                     media={mediaData}
//                     quality={PhotoQuality.HighRes}
//                     containerClass="cover-image"
//                 />
//                 <AlbumShareMenu
//                     album={album}
//                     closeShareMenu={() => setSharing(false)}
//                 />
//             </div>
//         </div>
//     )
// }
//
// function AlbumWrapper({
//     data,
//     index,
//     style,
// }: {
//     data: { albums: AlbumInfo[]; colCount: number }
//     index: number
//     style: React.CSSProperties
// }) {
//     const thisData = useMemo(() => {
//         const thisData = data.albums.slice(
//             index * data.colCount,
//             index * data.colCount + data.colCount
//         )
//
//         while (thisData.length < data.colCount) {
//             thisData.push({ id: '', name: '' } as AlbumInfo)
//         }
//
//         return thisData
//     }, [data, index])
//
//     return (
//         <div className="albums-row" style={style}>
//             {thisData.map((a, i) => {
//                 if (a.id !== '') {
//                     return <SingleAlbumCover key={a.id} album={a} />
//                 } else {
//                     return (
//                         <div key={`fake-album-${i}`} className="faux-album" />
//                     )
//                 }
//             })}
//         </div>
//     )
// }
//
// const ALBUM_WIDTH = 450
//
// export function AlbumScroller({ albums }: { albums: AlbumInfo[] }) {
//     const [containerRef, setContainerRef] = useState<HTMLDivElement>(null)
//     const containerSize = useResize(containerRef)
//
//     const colCount = Math.floor(containerSize.width / ALBUM_WIDTH)
//
//     return (
//         <div ref={setContainerRef} className="albums-container">
//             <FixedSizeList
//                 className="no-scrollbar"
//                 height={
//                     containerSize.height >= 21 ? containerSize.height - 21 : 0
//                 }
//                 width={containerSize.width}
//                 itemSize={480}
//                 itemCount={
//                     colCount !== 0 ? Math.ceil(albums.length / colCount) : 0
//                 }
//                 itemData={{
//                     albums: albums,
//                     colCount: colCount,
//                 }}
//             >
//                 {AlbumWrapper}
//             </FixedSizeList>
//         </div>
//     )
// }
