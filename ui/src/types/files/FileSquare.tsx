import { IconFolder } from '@tabler/icons-react'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { IconDisplay } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'

import { SelectedState, WeblensFile } from '@weblens/types/files/File'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from '@weblens/types/files/FileDragLogic'
import { MouseEvent, useState } from 'react'
import { useNavigate } from 'react-router-dom'

const FileGridVisual = ({ file }) => {
    return (
        <div className="w-full p-2 pb-0 aspect-square overflow-hidden">
            <div className="w-full h-full overflow-hidden rounded-md flex justify-center items-center">
                <IconDisplay file={file} allowMedia={true} />
            </div>
        </div>
    )
}

export const FileTextBox = ({
    itemTitle,
    doFolderIcon,
}: {
    itemTitle: string
    doFolderIcon: boolean
}) => {
    return (
        <div className="flex items-center justify-between w-[95%]">
            <p className="file-text">{itemTitle}</p>
            {doFolderIcon && (
                <IconFolder className="text-theme-text" stroke={3} />
            )}
        </div>
    )
}

export const FileSquare = ({ file }: { file: WeblensFile }) => {
    const [mouseDown, setMouseDown] = useState(null)

    const nav = useNavigate()

    const {
        draggingState,
        fbMode,
        shareId,
        menuMode,
        folderInfo,
        selected,
        setMoveDest,
        setHovering,
        setSelected,
        setDragging,
        setPresentationTarget,
        setMenu,
        clearSelected,
        setSelectedMoved,
    } = useFileBrowserStore()

    const selState = useFileBrowserStore((state) => {
        return state.filesMap.get(file?.Id())?.GetSelectedState()
    })

    return (
        <div
            className="weblens-file animate-fade"
            data-clickable={!draggingState || file.IsFolder()}
            data-hovering={selState & SelectedState.Hovering}
            data-in-range={(selState & SelectedState.InRange) >> 1}
            data-selected={(selState & SelectedState.Selected) >> 2}
            data-last-selected={(selState & SelectedState.LastSelected) >> 3}
            data-droppable={(selState & SelectedState.Droppable) >> 4}
            data-moved={(selState & SelectedState.Moved) >> 5}
            onMouseOver={(e: MouseEvent<HTMLDivElement>) =>
                handleMouseOver(
                    e,
                    file,
                    draggingState,
                    setHovering,
                    setMoveDest
                )
            }
            onMouseDown={(e) => {
                setMouseDown({ x: e.clientX, y: e.clientY })
            }}
            onMouseMove={(e) =>
                mouseMove(
                    e,
                    file,
                    draggingState,
                    mouseDown,
                    setSelected,
                    setDragging
                )
            }
            onClick={(e) => {
                e.stopPropagation()
                if (draggingState) {
                    return
                }
                setSelected([file.Id()])
            }}
            onDoubleClick={(e) =>
                visitFile(
                    e,
                    fbMode,
                    shareId,
                    file,
                    folderInfo.IsTrash(),
                    nav,
                    setPresentationTarget
                )
            }
            onContextMenu={(e) =>
                fileHandleContextMenu(e, menuMode, setMenu, file)
            }
            onMouseUp={(e) => {
                e.stopPropagation()
                return handleMouseUp(
                    file,
                    draggingState,
                    Array.from(selected.keys()),
                    setSelectedMoved,
                    clearSelected,
                    setMoveDest,
                    setDragging,
                    setMouseDown
                )
            }}
            onMouseLeave={() =>
                handleMouseLeave(
                    file,
                    draggingState,
                    setMoveDest,
                    setHovering,
                    mouseDown,
                    setMouseDown
                )
            }
        >
            <FileGridVisual file={file} />
            <div
                className="file-size-box"
                data-moved={(selState & SelectedState.Moved) >> 5}
            >
                <p className="file-size-text">{file.FormatSize()}</p>
            </div>
            <div className="file-text-container" style={{ height: '16%' }}>
                <FileTextBox
                    itemTitle={file.GetFilename()}
                    doFolderIcon={file.IsFolder() && file.GetContentId() !== ''}
                />
            </div>
        </div>
    )
}
