import React, { MouseEvent, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { SelectedState, WeblensFile } from './File'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from './FileDragLogic'
import { IconDisplay } from '../Pages/FileBrowser/FileBrowserMiscComponents'
import { useFileBrowserStore } from '../Pages/FileBrowser/FBStateControl'
import { useSessionStore } from '../components/UserInfo'

const FileGridVisual = ({ file }) => {
    return (
        <div className="w-full p-2 pb-0 aspect-square overflow-hidden">
            <div className="w-full h-full overflow-hidden rounded-md flex justify-center items-center">
                <IconDisplay file={file} allowMedia={true} />
            </div>
        </div>
    )
}

export const FileTextBox = ({ itemTitle }) => {
    return (
        <div className="file-text-container">
            <p className="p-2 truncate relative content-center text-[40cqh]">
                {itemTitle}
            </p>
        </div>
    )
}

export const FileSquare = ({ file }: { file: WeblensFile }) => {
    const [mouseDown, setMouseDown] = useState(null)
    const auth = useSessionStore((state) => state.auth)
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
            onMouseUp={() =>
                handleMouseUp(
                    file,
                    draggingState,
                    Array.from(selected.keys()),
                    auth,
                    setSelectedMoved,
                    clearSelected,
                    setMoveDest,
                    setDragging,
                    setMouseDown
                )
            }
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
                <p>{file.FormatSize()}</p>
            </div>
            <div className="flex relative items-center h-[16%] w-full">
                <FileTextBox itemTitle={file.GetFilename()} />
            </div>
        </div>
    )
}
