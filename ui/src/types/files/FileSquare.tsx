import { IconFolder } from '@tabler/icons-react'
import FileVisual from '@weblens/components/filebrowser/fileVisual'
import { HandleDrop } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import {
    fileHandleContextMenu,
    handleMouseLeave,
    handleMouseOver,
    handleMouseUp,
    mouseMove,
    visitFile,
} from '@weblens/types/files/FileDragLogic'
import filesStyle from '@weblens/types/files/filesStyle.module.scss'
import { MouseEvent, memo, useRef, useState } from 'react'

import { Coordinates, ErrorHandler } from '../Types'
import { DraggingStateT } from './FBTypes'
import WeblensFile, { SelectedState } from './File'

const FileGridVisual = ({
    file,
    allowMedia,
}: {
    file: WeblensFile
    allowMedia: boolean
}) => {
    return (
        <div className="group-data-moved:text-text-tertiary aspect-square w-full overflow-hidden p-2 pb-0">
            <div className="flex h-full w-full items-center justify-center overflow-hidden rounded-md">
                <FileVisual file={file} allowMedia={allowMedia} />
            </div>
        </div>
    )
}

const FileTextBox = ({
    file,
    selState,
    doFolderIcon,
}: {
    file: WeblensFile
    selState: SelectedState
    doFolderIcon: boolean
}) => {
    return (
        <div className="group-data-moved:text-text-tertiary relative flex h-full w-full items-center justify-between px-2">
            <span className={filesStyle.fileText}>{file.GetFilename()}</span>
            <div
                className={
                    'border-theme-primary flex max-h-6/10 w-max items-center rounded-md border p-1 text-[30cqh] select-none group-data-moved:border-none'
                }
                data-moved={(selState & SelectedState.Moved) >> 5 ? true : null}
            >
                <span className="text-[35cqh] text-current">
                    {file.FormatSize()}
                </span>
                {doFolderIcon && (
                    <IconFolder
                        className={filesStyle.fileSizeText + ' max-h-full'}
                        stroke={2}
                    />
                )}
            </div>
        </div>
    )
}

export const FileSquare = memo(
    ({ file, selState }: { file: WeblensFile; selState: SelectedState }) => {
        const [mouseDown, setMouseDown] = useState<Coordinates>(null)

        const draggingState = useFileBrowserStore(
            (state) => state.draggingState
        )
        const setMoveDest = useFileBrowserStore((state) => state.setMoveDest)
        const setHovering = useFileBrowserStore((state) => state.setHovering)
        const setSelected = useFileBrowserStore((state) => state.setSelected)
        const setDragging = useFileBrowserStore((state) => state.setDragging)
        const setPresentationTarget = useFileBrowserStore(
            (state) => state.setPresentationTarget
        )
        const setMenu = useFileBrowserStore((state) => state.setMenu)
        const clearSelected = useFileBrowserStore(
            (state) => state.clearSelected
        )
        const setSelectedMoved = useFileBrowserStore(
            (state) => state.setSelectedMoved
        )

        const fileRef = useRef<HTMLDivElement>()

        return (
            <div
                ref={fileRef}
                className={
                    'group bg-card-background-primary animate-fade-in data-selected:bg-card-background-selected hover:bg-background-tertiary data-droppable:bg-background-primary data-droppable:border-theme-primary border-theme-primary/0 relative m-1.5 flex aspect-[0.9] max-w-[400px] shrink basis-[250px] cursor-pointer flex-col rounded-md border-2 shadow transition select-none data-droppable:scale-101 data-moved:pointer-events-none'
                }
                data-clickable={!draggingState || file.IsFolder()}
                data-hovering={selState & SelectedState.Hovering}
                data-in-range={
                    (selState & SelectedState.InRange) >> 1 ? true : null
                }
                data-selected={
                    (selState & SelectedState.Selected) >> 2 ? true : null
                }
                data-last-selected={
                    (selState & SelectedState.LastSelected) >> 3 ? true : null
                }
                data-droppable={
                    (selState & SelectedState.Droppable) >> 4 ? true : null
                }
                data-moved={(selState & SelectedState.Moved) >> 5 ? true : null}
                data-dragging={draggingState}
                onMouseOver={(e: MouseEvent<HTMLDivElement>) =>
                    handleMouseOver(
                        e,
                        file,
                        draggingState,
                        setHovering,
                        setMoveDest,
                        setDragging
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
                    const sel = useFileBrowserStore.getState().selected

                    const exclusive = sel.size === 1 && sel.has(file.parentId)
                    setSelected([file.Id()], exclusive)
                }}
                onDoubleClick={(e) =>
                    visitFile(
                        e,
                        file,
                        useFileBrowserStore.getState().folderInfo.IsTrash(),
                        setPresentationTarget
                    )
                }
                onContextMenu={(e) => fileHandleContextMenu(e, setMenu, file)}
                onMouseUp={(e) => {
                    e.stopPropagation()
                    const sel = useFileBrowserStore.getState().selected
                    return handleMouseUp(
                        file,
                        draggingState,
                        Array.from(sel.keys()),
                        setSelectedMoved,
                        clearSelected,
                        setMoveDest,
                        setDragging,
                        setMouseDown,
                        useFileBrowserStore.getState().viewOpts.dirViewMode
                    )
                }}
                onMouseLeave={(e) =>
                    handleMouseLeave(
                        e,
                        file,
                        draggingState,
                        fileRef.current,
                        setMoveDest,
                        setHovering,
                        mouseDown,
                        setMouseDown
                    )
                }
                onDragEnter={(e) => {
                    handleMouseOver(
                        e,
                        file,
                        draggingState,
                        setHovering,
                        setMoveDest,
                        setDragging
                    )
                }}
                onDragOver={(e) => {
                    // https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                    e.preventDefault()
                }}
                onDragLeave={(e) => {
                    handleMouseLeave(
                        e,
                        file,
                        draggingState,
                        fileRef.current,
                        setMoveDest,
                        setHovering,
                        mouseDown,
                        setMouseDown
                    )
                }}
                onDrop={(e) => {
                    e.stopPropagation()
                    e.preventDefault()
                    if (
                        draggingState === DraggingStateT.ExternalDrag &&
                        file.GetSelectedState() & SelectedState.Droppable &&
                        file.IsFolder()
                    ) {
                        HandleDrop(
                            e.dataTransfer.items,
                            file.Id(),
                            false,
                            useFileBrowserStore.getState().shareId
                        ).catch(ErrorHandler)
                    }

                    setMoveDest('')
                    setDragging(DraggingStateT.NoDrag)
                    setHovering('')
                    file.SetSelected(SelectedState.Hovering, true)
                }}
            >
                <FileGridVisual
                    file={file}
                    allowMedia={!((selState & SelectedState.Moved) >> 5)}
                />
                <div
                    className={filesStyle.fileTextContainer}
                    style={{ height: '16%' }}
                >
                    <FileTextBox
                        file={file}
                        selState={selState}
                        doFolderIcon={
                            file.IsFolder() && file.GetContentId() !== ''
                        }
                    />
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.file !== next.file) {
            return false
        } else if (prev.selState !== next.selState) {
            return false
        }
        return true
    }
)
