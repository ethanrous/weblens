import {
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconServer,
    IconTrash,
    IconUsers,
} from '@tabler/icons-react'
import WeblensLoader from '@weblens/components/Loading.tsx'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensFileButton from '@weblens/lib/WeblensFileButton.tsx'
import { useResize } from '@weblens/lib/hooks'
import { DraggingStateT } from '@weblens/types/files/FBTypes'
import { FbMenuModeT } from '@weblens/types/files/File'
import User from '@weblens/types/user/User'
import { useRef } from 'react'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'

const EmptyIcon = ({ folderId, usr }: { folderId: string; usr: User }) => {
    if (folderId === usr.homeId) {
        return <IconHome size={500} className="text-nearly-invisible" />
    }
    if (folderId === usr.trashId) {
        return <IconTrash size={500} className="text-nearly-invisible" />
    }
    if (folderId === 'shared') {
        return <IconUsers size={500} className="text-nearly-invisible" />
    }
    if (folderId === 'EXTERNAL') {
        return <IconServer size={500} className="text-nearly-invisible" />
    }
    return <IconFolder size={500} className="text-nearly-invisible" />
}

function GetStartedCard() {
    const user = useSessionStore((state) => state.user)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const loading = useFileBrowserStore((state) => state.loading)
    const viewRef = useRef<HTMLDivElement>(null)
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
                <div className="wl-outline h-max w-max p-4">
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
            ref={viewRef}
            className="animate-fade z-3 m-auto flex h-max w-[50vw] min-w-[250px] items-center justify-center p-4"
        >
            <div className="flex h-fit w-max max-w-full flex-col items-center justify-center">
                <div className="pointer-events-none absolute -z-1 flex h-max max-w-full min-w-[200px] items-center p-30">
                    <EmptyIcon folderId={folderInfo.Id()} usr={user} />
                </div>

                <h4 className="z-10 h-max w-max select-none">
                    {`This folder ${folderInfo.IsPastFile() ? 'was' : 'is'} empty`}
                </h4>

                {folderInfo.IsModifiable() && (
                    <div
                        className="z-10 flex w-max max-w-full items-center gap-8 p-4"
                        style={{ flexDirection: doVertical ? 'column' : 'row' }}
                    >
                        <WeblensFileButton
                            folderId={folderInfo.Id()}
                            shareId={shareId}
                            multiple={true}
                            buttonProps={{ size: 'jumbo' }}
                        />
                        <WeblensButton
                            Left={IconFolderPlus}
                            squareSize={buttonSize}
                            size="jumbo"
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
