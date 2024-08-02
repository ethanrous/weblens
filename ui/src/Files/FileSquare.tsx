import React, { memo, MouseEvent, useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { SelectedState, WeblensFile } from './File'

import WeblensInput from '../components/WeblensInput'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from './FileDragLogic'
import { IconDisplay } from '../Pages/FileBrowser/FileBrowserMiscComponents'
import { handleRename } from '../Pages/FileBrowser/FileBrowserLogic'
import { useFileBrowserStore } from '../Pages/FileBrowser/FBStateControl'
import { useSessionStore } from '../components/UserInfo'

type TitleProps = {
    itemId: string
    itemTitle: string
    allowEditing: boolean
    blockFocus: (b: boolean) => void
    rename: (itemId: string, newName: string) => void
}

const FileGridVisual = ({ file }) => {
    return (
        <div className="w-full p-2 pb-0 aspect-square overflow-hidden">
            <div className="w-full h-full overflow-hidden rounded-md flex justify-center items-center">
                <IconDisplay file={file} allowMedia={true} />
            </div>
        </div>
    )
}

export const FileTextBox = memo(
    ({ itemId, itemTitle, blockFocus, rename }: TitleProps) => {
        const [renameVal, setRenameVal] = useState(itemTitle)

        const setEditingPlus = useCallback(
            (b: boolean) => {
                setRenameVal((cur) => {
                    if (cur === '') {
                        return itemTitle
                    } else {
                        return cur
                    }
                })
                blockFocus(b)
            },
            [itemTitle, blockFocus]
        )

        return (
            <div className="item-info-box select-none">
                <WeblensInput
                    subtle
                    fillWidth={false}
                    value={renameVal}
                    valueCallback={setRenameVal}
                    openInput={() => {
                        setEditingPlus(true)
                    }}
                    closeInput={() => {
                        setEditingPlus(false)
                        setRenameVal(itemTitle)
                        rename(itemId, '')
                    }}
                    onComplete={(v) => {
                        rename(itemId, v)
                        setEditingPlus(false)
                    }}
                />
            </div>
        )
    },
    (prev, next) => {
        if (prev.itemTitle !== next.itemTitle) {
            return false
        }
        return true
    }
)

export const FileSquare = ({ file }: { file: WeblensFile }) => {
    const [mouseDown, setMouseDown] = useState(null)
    const auth = useSessionStore((state) => state.auth)
    const nav = useNavigate()

    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const mode = useFileBrowserStore((state) => state.fbMode)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const menuMode = useFileBrowserStore((state) => state.menuMode)
    const selectedMap = useFileBrowserStore((state) => state.selected)

    const setBlockFocus = useFileBrowserStore((state) => state.setBlockFocus)
    const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
    const setHovering = useFileBrowserStore((state) => state.setHovering)
    const setSelected = useFileBrowserStore((state) => state.setSelected)
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const setPresentationTarget = useFileBrowserStore(
        (state) => state.setPresentationTarget
    )
    const setMenu = useFileBrowserStore((state) => state.setMenu)
    const clearSelected = useFileBrowserStore((state) => state.clearSelected)
    const setSelectedMoved = useFileBrowserStore(
        (state) => state.setSelectedMoved
    )

    const addLoading = useFileBrowserStore((state) => state.addLoading)
    const removeLoading = useFileBrowserStore((state) => state.removeLoading)

    const selected = useFileBrowserStore((state) =>
        state.filesMap.get(file.Id()).GetSelectedState()
    )

    return (
        <div
            className="weblens-file animate-fade"
            data-clickable={!draggingState || file.IsFolder()}
            data-hovering={selected & SelectedState.Hovering}
            data-in-range={(selected & SelectedState.InRange) >> 1}
            data-selected={(selected & SelectedState.Selected) >> 2}
            data-last-selected={(selected & SelectedState.LastSelected) >> 3}
            data-droppable={(selected & SelectedState.Droppable) >> 4}
            data-moved={(selected & SelectedState.Moved) >> 5}
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
                visitFile(e, mode, shareId, file, nav, setPresentationTarget)
            }
            onContextMenu={(e) =>
                fileHandleContextMenu(e, menuMode, setMenu, file)
            }
            onMouseUp={() =>
                handleMouseUp(
                    file,
                    draggingState,
                    Array.from(selectedMap.keys()),
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
                data-moved={(selected & SelectedState.Moved) >> 5}
            >
                <p>{file.FormatSize()}</p>
            </div>
            <div className="flex h-[16%] w-full">
                <FileTextBox
                    itemId={file.Id()}
                    itemTitle={file.GetFilename()}
                    allowEditing={file.IsModifiable()}
                    blockFocus={setBlockFocus}
                    rename={(itemId, newName) =>
                        handleRename(
                            itemId,
                            newName,
                            addLoading,
                            removeLoading,
                            auth
                        )
                    }
                />
            </div>
        </div>
    )
}
