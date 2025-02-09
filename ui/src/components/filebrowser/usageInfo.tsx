import {
    IconFiles,
    IconFolder,
    IconFolders,
    IconHome,
} from '@tabler/icons-react'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useResize } from '@weblens/components/hooks'
import theme from '@weblens/components/theme.module.scss'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { humanFileSize } from '@weblens/util'
import { useMemo, useState } from 'react'

import { FbModeT, useFileBrowserStore } from '../../store/FBStateControl'

function UsageInfo() {
    const [box, setBox] = useState<HTMLDivElement>(null)
    const size = useResize(box)

    const user = useSessionStore((state) => state.user)

    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    let homeSize = useFileBrowserStore((state) => state.homeDirSize)
    const trashSize = useFileBrowserStore((state) => state.trashDirSize)
    const selected = useFileBrowserStore((state) => state.selected)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const pastTime = useFileBrowserStore((state) => state.pastTime)

    const selectedLength = selected.size

    const mode = useFileBrowserStore((state) => state.fbMode)

    const { selectedFileSize, selectedFolderCount, selectedFileCount } =
        useMemo(() => {
            let selectedFileSize = 0
            let selectedFolderCount = 0
            let selectedFileCount = 0
            Array.from(selected.keys()).forEach((fileId) => {
                const f = filesMap.get(fileId)
                if (!f) {
                    return
                }
                if (f.IsFolder()) {
                    selectedFolderCount++
                } else {
                    selectedFileCount++
                }
                selectedFileSize += f.size || 0
            })

            return { selectedFileSize, selectedFolderCount, selectedFileCount }
        }, [selectedLength])

    let displaySize = folderInfo?.GetSize() || 0

    if (folderInfo?.Id() !== user.trashId) {
        homeSize = homeSize - trashSize
    }

    if (homeSize < displaySize) {
        displaySize = homeSize
    }

    const doGlobalSize = selectedLength === 0 && mode !== FbModeT.share

    let usagePercent = 100
    if (pastTime.getTime() === 0) {
        usagePercent = doGlobalSize
            ? (displaySize / homeSize) * 100
            : (selectedFileSize / displaySize) * 100
        if (!usagePercent || (selectedLength !== 0 && displaySize === 0)) {
            usagePercent = 0
        }
    }

    const miniMode = size.width !== -1 && size.width < 100

    let StartIcon = doGlobalSize ? IconFolder : IconFiles
    let EndIcon = doGlobalSize ? IconHome : IconFolder
    if (miniMode) {
        ;[StartIcon, EndIcon] = [EndIcon, StartIcon]
    }

    let startSize = doGlobalSize
        ? humanFileSize(displaySize).join(' ')
        : humanFileSize(selectedFileSize).join(' ')

    let endSize = doGlobalSize
        ? humanFileSize(homeSize).join(' ')
        : humanFileSize(displaySize).join(' ')

    if (
        pastTime.getTime() !== 0 ||
        folderInfo?.Id() === 'shared' ||
        user === null ||
        homeSize === -1 ||
        trashSize === -1
    ) {
        startSize = '--'
        endSize = '--'
    }

    return (
        <div
            ref={setBox}
            className="flex flex-col h-max w-full gap-3 mb-2"
            style={{
                alignItems: miniMode ? 'center' : 'flex-start',
            }}
        >
            {miniMode && <StartIcon className={theme['background-icon']} />}
            <div
                className="relative h-max w-max"
                style={{
                    height: miniMode ? 100 : 20,
                    width: miniMode ? 20 : '100%',
                }}
            >
                <WeblensProgress
                    key={miniMode ? 'usage-vertical' : 'usage-horizontal'}
                    value={usagePercent}
                    orientation={miniMode ? 'vertical' : 'horizontal'}
                />
            </div>
            <div
                className="flex flex-row h-max justify-between items-center"
                style={{
                    width: miniMode ? 'max-content' : '98%',
                }}
            >
                {folderInfo?.Id() !== 'shared' && !miniMode && (
                    <div className="flex flex-row items-center">
                        {<StartIcon className={theme['background-icon']} />}
                        <p
                            className="select-none p-1"
                            style={{
                                display: miniMode ? 'none' : 'block',
                            }}
                        >
                            {startSize}
                        </p>
                    </div>
                )}
                <div className="flex flex-row justify-end w-max items-center">
                    <p
                        className="select-none p-1"
                        style={{
                            display: miniMode ? 'none' : 'block',
                        }}
                    >
                        {endSize}
                    </p>
                    {<EndIcon className={theme['background-icon']} />}
                </div>
            </div>
            <div
                className="flex flex-row h-max justify-between items-center w-full bg-[--wl-barely-visible] rounded-lg p-2"
                style={{
                    display: selectedLength > 0 ? 'flex' : 'none',
                    flexDirection: miniMode ? 'column' : 'row',
                }}
            >
                <div className="flex h-max items-center text-[--wl-text-color] w-min">
                    <IconFiles />
                    <p className="select-none p-1">{selectedFileCount}</p>
                </div>
                <div className="flex h-max items-center text-[--wl-text-color] w-min">
                    <IconFolders />
                    <p className="select-none p-1">{selectedFolderCount}</p>
                </div>
            </div>
        </div>
    )
}

export default UsageInfo
