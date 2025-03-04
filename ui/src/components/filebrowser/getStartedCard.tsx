import { Divider, FileButton } from '@mantine/core'
import {
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconServer,
    IconTrash,
    IconUpload,
    IconUsers,
} from '@tabler/icons-react'
import WeblensLoader from '@weblens/components/Loading'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useResize } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { HandleUploadButton } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { ErrorHandler } from '@weblens/types/Types'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { FbMenuModeT } from '@weblens/types/files/File'
import User from '@weblens/types/user/User'
import { useState } from 'react'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'

const EmptyIcon = ({ folderId, usr }: { folderId: string; usr: User }) => {
    if (folderId === usr.homeId) {
        return (
            <IconHome size={500} className="text-wl-color-nearly-invisible" />
        )
    }
    if (folderId === usr.trashId) {
        return (
            <IconTrash size={500} className="text-wl-color-nearly-invisible" />
        )
    }
    if (folderId === 'shared') {
        return (
            <IconUsers size={500} className="text-wl-color-nearly-invisible" />
        )
    }
    if (folderId === 'EXTERNAL') {
        return (
            <IconServer size={500} className="text-wl-color-nearly-invisible" />
        )
    }
    return <IconFolder size={500} className="text-wl-color-nearly-invisible" />
}

function GetStartedCard() {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const loading = useFileBrowserStore((state) => state.loading)
    const [viewRef, setViewRef] = useState<HTMLDivElement>()
    const size = useResize(viewRef)

    const setMenu = useFileBrowserStore((state) => state.setMenu)

    if (draggingState === DraggingStateT.ExternalDrag) {
        return null
    }

    if (loading && loading.length !== 0) {
        return <WeblensLoader />
    }

    if (!folderInfo && mode === FbModeT.share) {
        return (
            <div className="m-auto flex items-center justify-center p-2">
                <div className="h-max w-max p-4 wl-outline">
                    <h4>You have no files shared with you</h4>
                </div>
            </div>
        )
    } else if (!folderInfo) {
        return null
    }

    const doVertical = size.width < 500

    let buttonSize = 128
    if (doVertical) {
        buttonSize = 64
    }

    return (
        <div
            ref={setViewRef}
            className="z-[3] m-auto flex h-max w-[50vw] min-w-[250px] animate-fade items-center justify-center p-4"
        >
            <div className="flex h-fit w-max max-w-full flex-col items-center justify-center">
                <div className="p-30 -z-1 pointer-events-none absolute flex h-max min-w-[200px] max-w-full items-center">
                    <EmptyIcon folderId={folderInfo.Id()} usr={user} />
                </div>

                <p className="z-10 h-max w-max select-none text-2xl">
                    {`This folder ${folderInfo.IsPastFile() ? 'was' : 'is'} empty`}
                </p>

                {folderInfo.IsModifiable() && (
                    <div
                        className="w-350 z-10 flex max-w-full items-center p-5"
                        style={{ flexDirection: doVertical ? 'column' : 'row' }}
                    >
                        <FileButton
                            onChange={(files) => {
                                HandleUploadButton(
                                    files,
                                    folderInfo.Id(),
                                    false,
                                    shareId
                                ).catch(ErrorHandler)
                            }}
                            accept="file"
                            multiple
                        >
                            {(props) => {
                                return (
                                    <WeblensButton
                                        subtle
                                        Left={IconUpload}
                                        squareSize={buttonSize}
                                        tooltip="Upload"
                                        onClick={props.onClick}
                                    />
                                )
                            }}
                        </FileButton>
                        <Divider
                            orientation={doVertical ? 'horizontal' : 'vertical'}
                            m={30 * (buttonSize / 128)}
                        />

                        <WeblensButton
                            Left={IconFolderPlus}
                            squareSize={buttonSize}
                            subtle
                            tooltip="New Folder"
                            onClick={(e) => {
                                setMenu({
                                    menuPos: { x: e.clientX, y: e.clientY },
                                    menuState: FbMenuModeT.NameFolder,
                                })
                            }}
                        />
                    </div>
                )}
            </div>
        </div>
    )
}

export default GetStartedCard
