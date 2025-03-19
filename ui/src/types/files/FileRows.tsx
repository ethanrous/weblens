import FileVisual from '@weblens/components/filebrowser/fileVisual'
import GetStartedCard from '@weblens/components/filebrowser/getStartedCard'
import {
    HandleDrop,
    historyDateTime,
} from '@weblens/pages/FileBrowser/FileBrowserLogic'
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
import { useResize } from '@weblens/lib/hooks'
import { CSSProperties, MouseEvent, useRef, useState } from 'react'
import { FixedSizeList as WindowList } from 'react-window'

import { Coordinates, ErrorHandler } from '../Types'
import { DraggingStateT } from './FBTypes'
import WeblensFile, { SelectedState } from './File'

function FileRow({
    data,
    index,
    style,
}: {
    data: WeblensFile[]
    index: number
    style: CSSProperties
}) {
    const file = data[index]
    const fileRef = useRef<HTMLDivElement>()

    const [mouseDown, setMouseDown] = useState<Coordinates>(null)

    const draggingState = useFileBrowserStore((state) => state.draggingState)
    const viewOpts = useFileBrowserStore((state) => state.viewOpts)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const selected = useFileBrowserStore((state) => state.selected)

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

    const selState = useFileBrowserStore((state) =>
        state.filesMap.get(file.Id()).GetSelectedState()
    )

    return (
        <div style={{ ...style, padding: 4 }}>
            <div
                ref={fileRef}
                className={filesStyle.weblensFile + ' animate-fade-short'}
                data-row={true}
                data-clickable={!draggingState || file.IsFolder()}
                data-hovering={selState & SelectedState.Hovering}
                data-in-range={(selState & SelectedState.InRange) >> 1}
                data-selected={(selState & SelectedState.Selected) >> 2}
                data-last-selected={
                    (selState & SelectedState.LastSelected) >> 3
                }
                data-droppable={(selState & SelectedState.Droppable) >> 4}
                data-moved={(selState & SelectedState.Moved) >> 5}
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
                    setSelected([file.Id()])
                }}
                onDoubleClick={(e) =>
                    visitFile(
                        e,
                        file,
                        folderInfo.IsTrash(),
                        setPresentationTarget
                    )
                }
                onContextMenu={(e) => fileHandleContextMenu(e, setMenu, file)}
                onMouseUp={() => {
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
            >
                <div className={filesStyle.fileRowBox}>
                    <div className="m-1 flex aspect-square h-full shrink-0 items-center justify-center overflow-hidden rounded-sm">
                        <FileVisual file={file} allowMedia={true} />
                    </div>
                    <div className="flex h-full grow flex-col">
                        <div className={filesStyle.fileTextContainer}>
                            <h1 className={filesStyle.fileText}>
                                {file.GetFilename()}
                            </h1>
                        </div>
                        <p className="selectable-text w-max pl-1 text-xs">
                            {historyDateTime(file.GetModified().getTime())}
                        </p>
                    </div>
                    <div
                        className={filesStyle.fileSizeBox}
                        data-moved={(selState & SelectedState.Moved) >> 5}
                    >
                        <p className={filesStyle.fileSizeText}>
                            {file.FormatSize()}
                        </p>
                    </div>
                </div>
                <div className="flex h-full flex-col"></div>
            </div>
        </div>
    )
}

export function FileRows({ files }: { files: WeblensFile[] }) {
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const size = useResize(boxRef)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const dragState = useFileBrowserStore((state) => state.draggingState)
    const shareId = useFileBrowserStore((state) => state.shareId)

    const setDragging = useFileBrowserStore((state) => state.setDragging)

    return (
        <div
            ref={setBoxRef}
            className={filesStyle.fileRows}
            data-droppable={Boolean(
                moveDest === folderInfo?.Id() &&
                    folderInfo?.modifiable &&
                    dragState === DraggingStateT.ExternalDrag
            )}
            data-bad-drop={Boolean(
                moveDest === folderInfo?.Id() &&
                    !folderInfo?.modifiable &&
                    dragState === DraggingStateT.ExternalDrag
            )}
            onDragOver={(e) => {
                // https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                if (dragState === DraggingStateT.ExternalDrag) {
                    e.preventDefault()
                }
            }}
            onDrop={(e) => {
                if (dragState !== DraggingStateT.ExternalDrag) {
                    return
                }

                e.preventDefault()
                if (folderInfo?.modifiable) {
                    HandleDrop(
                        e.dataTransfer.items,
                        folderInfo.Id(),
                        false,
                        shareId
                    ).catch(ErrorHandler)
                }

                setDragging(DraggingStateT.NoDrag)
            }}
        >
            {files.length === 0 && <GetStartedCard />}
            {files.length !== 0 && (
                <WindowList
                    height={size.height}
                    width={size.width - 4}
                    itemSize={70}
                    itemCount={files.length}
                    itemData={files}
                    overscanCount={10}
                >
                    {FileRow}
                </WindowList>
            )}
        </div>
    )
}
