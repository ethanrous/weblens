import { IconFolder } from '@tabler/icons-react'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { HandleDrop } from '@weblens/pages/FileBrowser/FileBrowserLogic'
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
import filesStyle from '@weblens/types/files/filesStyle.module.scss'
import { MouseEvent, useMemo, useRef, useState } from 'react'

import { Coordinates, ErrorHandler } from '../Types'
import { DraggingStateT } from './FBTypes'

const FileGridVisual = ({
    file,
    allowMedia,
}: {
    file: WeblensFile
    allowMedia: boolean
}) => {
    return (
        <div className="w-full p-2 pb-0 aspect-square overflow-hidden">
            <div className="w-full h-full overflow-hidden rounded-md flex justify-center items-center">
                <IconDisplay file={file} allowMedia={allowMedia} />
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
        <div className="flex items-center justify-between px-2 relative w-full h-full">
            <p className={filesStyle['file-text']}>{file.GetFilename()}</p>
            <div
                className={filesStyle['file-size-box']}
                data-moved={(selState & SelectedState.Moved) >> 5}
            >
                <h4 className={filesStyle['file-size-text']}>
                    {file.FormatSize()}
                </h4>
                {doFolderIcon && (
                    <IconFolder
                        className={filesStyle['file-size-text'] + " max-h-full"}
                        stroke={2}
                    />
                )}
            </div>
        </div>
    )
}

export const FileSquare = ({ file }: { file: WeblensFile }) => {
    const [mouseDown, setMouseDown] = useState<Coordinates>(null)

    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const hoveringId = useFileBrowserStore((state) => state.hoveringId)
    const holdingShift = useFileBrowserStore((state) => state.holdingShift)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const selected = useFileBrowserStore((state) => state.selected)
    const shareId = useFileBrowserStore((state) => state.shareId)
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

    const selState = useMemo(() => {
        return filesMap.get(file?.Id())?.GetSelectedState()
    }, [file, selected, hoveringId, holdingShift])
    const fileRef = useRef<HTMLDivElement>()

    return (
        <div
            ref={fileRef}
            className={filesStyle['weblens-file'] + ' animate-fade'}
            data-clickable={!draggingState || file.IsFolder()}
            data-hovering={selState & SelectedState.Hovering}
            data-in-range={(selState & SelectedState.InRange) >> 1}
            data-selected={(selState & SelectedState.Selected) >> 2}
            data-last-selected={(selState & SelectedState.LastSelected) >> 3}
            data-droppable={(selState & SelectedState.Droppable) >> 4}
            data-moved={(selState & SelectedState.Moved) >> 5}
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

                const exclusive =
                    selected.size === 1 && selected.has(file.parentId)
                setSelected([file.Id()], exclusive)
            }}
            onDoubleClick={(e) =>
                visitFile(e, file, folderInfo?.IsTrash(), setPresentationTarget)
            }
            onContextMenu={(e) => fileHandleContextMenu(e, setMenu, file)}
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
                    setMouseDown,
                    viewOpts.dirViewMode
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
                        [],
                        false,
                        shareId
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
                className={filesStyle['file-text-container']}
                style={{ height: '16%' }}
            >
                <FileTextBox
                    file={file}
                    selState={selState}
                    doFolderIcon={file.IsFolder() && file.GetContentId() !== ''}
                />
            </div>
        </div>
    )
}
