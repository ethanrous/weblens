import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { historyDate } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import {
    GetStartedCard,
    IconDisplay,
} from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
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
import { useResize } from 'components/hooks'
import { CSSProperties, MouseEvent, useRef, useState } from 'react'
import { FixedSizeList as WindowList } from 'react-window'

import { Coordinates } from '../Types'

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
                className={filesStyle['weblens-file'] + ' animate-fade-short'}
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
                <div className={filesStyle['file-row-box']}>
                    <div className="flex shrink-0 h-full aspect-square rounded overflow-hidden m-1 justify-center items-center">
                        <IconDisplay file={file} allowMedia={true} />
                    </div>
                    <div className="flex flex-col h-full grow">
                        <div className={filesStyle['file-text-container']}>
                            <h1 className={filesStyle['file-text']}>
                                {file.GetFilename()}
                            </h1>
                        </div>
                        <p className="selectable-text w-max text-xs pl-1">
                            {historyDate(file.GetModified().getTime())}
                        </p>
                    </div>
                    <div
                        className={filesStyle['file-size-box']}
                        data-moved={(selState & SelectedState.Moved) >> 5}
                    >
                        <p className={filesStyle['file-size-text']}>
                            {file.FormatSize()}
                        </p>
                    </div>
                </div>
                <div className="flex flex-col h-full"></div>
            </div>
        </div>
    )
}

export function FileRows({ files }: { files: WeblensFile[] }) {
    const [boxRef, setBoxRef] = useState<HTMLDivElement>()
    const size = useResize(boxRef)

    return (
        <div ref={setBoxRef} className={filesStyle['file-rows']}>
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
