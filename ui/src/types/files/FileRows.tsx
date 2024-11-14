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
import { useResize } from 'components/hooks'
import { CSSProperties, MouseEvent, useState } from 'react'
import { FixedSizeList as WindowList } from 'react-window'
import '@weblens/types/files/filesStyle.scss'
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

    const [mouseDown, setMouseDown] = useState<Coordinates>(null)

    const {
        draggingState,
        viewOpts,
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

    const selState = useFileBrowserStore((state) =>
        state.filesMap.get(file.Id()).GetSelectedState()
    )

    return (
        <div style={{ ...style, padding: 4 }}>
            <div
                className="weblens-file animate-fade-short"
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
                <div className="flex h-full w-full items-center">
                    <div className="flex shrink-0 h-full aspect-square rounded overflow-hidden m-1 justify-center items-center">
                        <IconDisplay file={file} allowMedia={true} />
                    </div>
                    <div className="flex flex-col h-full">
                        <div className="file-text-container">
                            <h1 className="file-text">{file.GetFilename()}</h1>
                        </div>
                        <p className="selectable-text w-max text-xs pl-1">
                            {historyDate(file.GetModified().getTime())}
                        </p>
                    </div>
                    <div
                        className="file-size-box"
                        data-moved={(selState & SelectedState.Moved) >> 5}
                    >
                        <p className="file-size-text">{file.FormatSize()}</p>
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

    if (files.length === 0) {
        return <GetStartedCard />
    }

    return (
        <div ref={setBoxRef} className="file-rows no-scrollbar">
            <WindowList
                height={size.height}
                width={size.width - 4}
                itemSize={70}
                itemCount={files.length}
                itemData={files}
                overscanCount={25}
            >
                {FileRow}
            </WindowList>
        </div>
    )
}
