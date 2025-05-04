import { FileApi, GetFolderData } from '@weblens/api/FileBrowserApi'
import SharesApi from '@weblens/api/SharesApi'
import { useSubscribe as useFolderSubscribe } from '@weblens/api/Websocket'
import HeaderBar from '@weblens/components/HeaderBar'
import { PresentationFile } from '@weblens/components/Presentation'
import { useSessionStore } from '@weblens/components/UserInfo'
import { FileContextMenu } from '@weblens/components/filebrowser/FileMenu'
import DirectoryView from '@weblens/components/filebrowser/directoryView'
import FBSidebar from '@weblens/components/filebrowser/filebrowserSidebar'
import PasteDialogue from '@weblens/components/filebrowser/pasteDialogue'
import SearchDialogue from '@weblens/components/filebrowser/searchDialogue'
import WebsocketStatusDot from '@weblens/components/filebrowser/websocketStatus'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensFile from '@weblens/types/files/File'
import { goToFile } from '@weblens/types/files/FileDragLogic'
import { useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'
import { DraggingCounter } from './DropSpot'
import {
    filenameFromPath,
    getRealId,
    useKeyDownFileBrowser,
    usePaste,
} from './FileBrowserLogic'
import { DirViewModeT } from './FileBrowserTypes'

function useSearch() {
    const { search } = useLocation()
    const q = new URLSearchParams(search)
    return useCallback(
        (s: string) => {
            const r = q.get(s)
            if (!r) {
                return ''
            }
            return r
        },
        [q]
    )
}

function FileBrowser() {
    const urlPath = useParams()['*']
    const jumpTo = window.location.hash.substring(1)
    const query = useSearch()
    const user = useSessionStore((state) => state.user)
    const nav = useNavigate()

    useEffect(() => {
        useFileBrowserStore.getState().setNav(nav)
    }, [nav])

    const [filesFetchErr, setFilesFetchErr] = useState(0)

    const {
        viewOpts,
        blockFocus,
        filesMap,
        filesLists,
        folderInfo,
        presentingId,
        isSearching,
        fbMode,
        contentId,
        shareId,
        pastTime,
        addLoading,
        removeLoading,
        setLocationState,
        setSelected,
        clearSelected,
        setFilesData,
    } = useFileBrowserStore()

    useEffect(() => {
        localStorage.setItem('fbViewOpts', JSON.stringify(viewOpts))
    }, [viewOpts])

    const past = query('past')

    useEffect(() => {
        if (!user) {
            return
        }

        let mode: FbModeT = 0
        let contentId: string = ''
        let shareId: string = ''
        const splitPath: string[] = urlPath
            .split('/')
            .filter((s) => s.length !== 0)

        if (splitPath.length === 0) {
            return
        }

        if (splitPath[0] === 'share') {
            mode = FbModeT.share
            shareId = splitPath[1]
            contentId = splitPath[2]
        } else if (splitPath[0] === 'shared') {
            mode = FbModeT.share
            shareId = ''
            contentId = ''
        } else if (splitPath[0] === 'external') {
            mode = FbModeT.external
            contentId = splitPath[1]
        } else if (splitPath[0] === 'stats') {
            mode = FbModeT.stats
            contentId = splitPath[1]
        } else if (splitPath[0] === 'search') {
            mode = FbModeT.search
            contentId = splitPath[1]
        } else {
            mode = FbModeT.default
            contentId = splitPath[0]
        }

        const pastTime: Date = past ? new Date(past) : null

        if (mode === FbModeT.share && shareId && !contentId) {
            SharesApi.getFileShare(shareId)
                .then((res) => {
                    nav(`/files/share/${shareId}/${res.data.fileId}`)
                })
                .catch(ErrorHandler)
        } else {
            contentId = getRealId(contentId, mode, user)
            setLocationState({ contentId, mode, shareId, pastTime, jumpTo })
            // removeLoading('files')
        }
    }, [urlPath, user, past, jumpTo])

    const { readyState } = useFolderSubscribe()

    useKeyDownFileBrowser()

    // Hook to handle uploading images from the clipboard
    usePaste(contentId, user, blockFocus)

    // Reset most of the state when we change folders
    useEffect(() => {
        if (contentId === null) {
            console.debug('Content ID is null, refusing to sync state')
            return
        }

        const syncState = async () => {
            setFilesFetchErr(0)

            if (!urlPath) {
                nav('/files/home', { replace: true })
            }

            if (urlPath === user?.homeId) {
                const redirect = '/files/home' + window.location.search
                nav(redirect, { replace: true })
            }

            // If we're not ready, leave
            if (fbMode == FbModeT.unset || !user) {
                console.debug(
                    'Not ready to sync state. Mode: ',
                    fbMode,
                    'User:',
                    user
                )
                return
            }

            if (!user.isLoggedIn && fbMode !== FbModeT.share) {
                console.debug('Going to login')
                nav('/login', { state: { returnTo: window.location.pathname } })
            }

            if (viewOpts.dirViewMode !== DirViewModeT.Columns) {
                clearSelected()
            }

            const folder = filesMap.get(contentId)
            if (
                folder &&
                (pastTime.getTime() === 0 || folder.modifyDate === pastTime) &&
                (folder.GetFetching() ||
                    (folder.modifiable !== undefined &&
                        folder.childrenIds &&
                        folder.childrenIds.filter((f) => f !== user.trashId)
                            .length === filesLists.get(folder.Id())?.length))
            ) {
                console.debug('Exiting sync state early')
                if (folder.Id() !== folderInfo.Id()) {
                    setFilesData({
                        selfInfo: folder,
                    })
                }
                return
            }

            folder?.SetFetching(true)

            const fileData = await GetFolderData(
                contentId,
                fbMode,
                shareId,
                pastTime
            ).catch((r: number) => {
                setFilesFetchErr(r)
            })

            // If request comes back after we have already navigated away, do nothing
            if (useFileBrowserStore.getState().contentId !== contentId) {
                console.error("Content ID don't match")
                return
            }

            if (fileData) {
                if (
                    fbMode === FbModeT.share &&
                    fileData.self?.owner == user.username
                ) {
                    nav(`/files/${fileData.self.id}`, { replace: true })
                    return
                }

                console.debug('Setting main files data', fileData)
                setFilesData({
                    selfInfo: fileData.self,
                    childrenInfo: fileData.children,
                    parentsInfo: fileData.parents,
                    mediaData: fileData.medias,
                })

                document.title =
                    filenameFromPath(fileData.self.portablePath).nameText +
                    ' - Weblens'
                folder?.SetFetching(false)
            }

            if (
                (jumpTo || viewOpts.dirViewMode === DirViewModeT.Columns) &&
                (fbMode !== FbModeT.share || contentId) &&
                useFileBrowserStore.getState().selected.size === 0
            ) {
                setSelected([jumpTo ? jumpTo : contentId], true)
            }
        }

        addLoading('files')
        syncState()
            .catch((e: number) => {
                console.error(e)
                setFilesFetchErr(e)
            })
            .finally(() => removeLoading('files'))
    }, [user, contentId, shareId, fbMode, pastTime, jumpTo])

    useEffect(() => {
        const selectedSize = useFileBrowserStore.getState().selected.size

        if (
            viewOpts.dirViewMode !== DirViewModeT.Columns &&
            selectedSize === 1
        ) {
            clearSelected()
        }

        if (
            (jumpTo || viewOpts.dirViewMode === DirViewModeT.Columns) &&
            (fbMode !== FbModeT.share || contentId) &&
            selectedSize === 0
        ) {
            setSelected([jumpTo ? jumpTo : contentId], true)
        }
    }, [viewOpts.dirViewMode])

    const searchVisitFunc = (loc: string) => {
        FileApi.getFile(loc)
            .then((f) => {
                if (!f.data) {
                    console.error('Could not find file to nav to')
                    return
                }

                goToFile(new WeblensFile(f.data), true)
            })
            .catch(ErrorHandler)
    }

    return (
        <div className="flex h-screen flex-col">
            <HeaderBar />
            <DraggingCounter />
            <PresentationFile file={filesMap.get(presentingId)} />
            <PasteDialogue />
            {isSearching && (
                <div className="absolute z-40 flex h-screen w-screen items-center justify-center bg-[#00000088] px-[30%] py-[10%] backdrop-blur-xs">
                    <SearchDialogue text={''} visitFunc={searchVisitFunc} />
                </div>
            )}
            <FileContextMenu />
            <div className="absolute bottom-1 left-1">
                <WebsocketStatusDot ready={readyState} />
            </div>
            <div className="flex h-[90vh] grow flex-row items-start">
                <FBSidebar />
                <DirectoryView
                    filesError={filesFetchErr}
                    setFilesError={setFilesFetchErr}
                    searchFilter={''}
                />
            </div>
        </div>
    )
}

export default FileBrowser
