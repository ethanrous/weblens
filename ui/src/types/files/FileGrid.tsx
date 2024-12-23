import WeblensLoader from '@weblens/components/Loading'
import { HandleDrop } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { GetStartedCard } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { WeblensFile } from '@weblens/types/files/File'
import { FileSquare } from '@weblens/types/files/FileSquare'
import filesStyle from '@weblens/types/files/filesStyle.module.scss'
import { useResize } from 'components/hooks'
import { CSSProperties, useEffect, useMemo, useRef, useState } from 'react'
import { FixedSizeGrid as Grid } from 'react-window'

import { ErrorHandler } from '../Types'
import { DraggingStateT } from './FBTypes'

type GridDataProps = {
    files: WeblensFile[]
    numCols: number
}

function SquareWrapper({
    data,
    rowIndex,
    columnIndex,
    style,
}: {
    data: GridDataProps
    rowIndex: number
    columnIndex: number
    style: CSSProperties
}) {
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const hoveringId = useFileBrowserStore((state) => state.hoveringId)
    const holdingShift = useFileBrowserStore((state) => state.holdingShift)
    const selected = useFileBrowserStore((state) => state.selected)

    const absIndex = rowIndex * data.numCols + columnIndex
    const file = data.files[absIndex]

    const selState = useMemo(() => {
        if (!file) {
            return 0
        }
        return filesMap.get(file?.Id())?.GetSelectedState()
    }, [file, hoveringId, holdingShift, filesMap, selected])

    if (!file) {
        return null
    }

    if (!data || rowIndex === undefined) {
        return null
    }

    return (
        <div style={style}>
            <FileSquare file={file} selState={selState} />
        </div>
    )
}

function FileGrid({ files }: { files: WeblensFile[] }) {
    const jumpTo = useFileBrowserStore((state) => state.jumpTo)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    const gridRef = useRef<Grid>()
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const [didScroll, setDidScroll] = useState<boolean>()
    const [lastSeen, setLastSeen] = useState<{
        file: WeblensFile
        width: number
    }>()
    const [timeoutId, setTimeoutId] = useState<NodeJS.Timeout>()
    const size = useResize(containerRef)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const moveDest = useFileBrowserStore((state) => state.moveDest)
    const dragState = useFileBrowserStore((state) => state.draggingState)
    const loading = useFileBrowserStore((state) => state.loading)
    const setDragging = useFileBrowserStore((state) => state.setDragging)
    const numCols = Math.max(Math.floor(size.width / 250), 2)

    const squareSize = (size.width / numCols) * 1.15
    const margin = 8
    const rowHeight = squareSize + margin
    const filteredFiles = useMemo(() => {
        const filteredFiles = files.filter(
            (file) => file.parentId === folderInfo.Id()
        )
        if (filteredFiles.length === 0) {
            return []
        }
        return filteredFiles
    }, [files])

    useEffect(() => {
        if (lastSeen?.file) {
            if (timeoutId) {
                clearTimeout(timeoutId)
            }
            setTimeoutId(
                setTimeout(() => {
                    if (!gridRef.current) {
                        return
                    }
                    gridRef.current.scrollToItem({
                        align: 'smart',
                        rowIndex: Math.floor(
                            lastSeen.file.GetIndex() / numCols
                        ),
                    })
                }, 100)
            )
        }
    }, [size.width])

    const isLoading = loading.includes('files')

    return (
        <div
            ref={setContainerRef}
            className={filesStyle['files-grid']}
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
            {size.width !== -1 && (
                <div className="flex relative w-full h-full items-center">
                    {isLoading && (
                        <div className="m-auto p-2">
                            <WeblensLoader />
                        </div>
                    )}
                    {!isLoading && files.length === 0 && <GetStartedCard />}
                    {!isLoading && files.length !== 0 && (
                        <Grid
                            className="no-scrollbar outline-0"
                            ref={gridRef}
                            columnCount={numCols}
                            itemData={{ files: filteredFiles, numCols }}
                            height={size.height}
                            width={size.width}
                            rowCount={Math.ceil(filteredFiles.length / numCols)}
                            columnWidth={size.width / numCols}
                            rowHeight={rowHeight}
                            overscanRowCount={8}
                            onScroll={({ scrollTop }) => {
                                if (
                                    lastSeen &&
                                    lastSeen.width !== size.width &&
                                    lastSeen.width !== 0
                                ) {
                                    setLastSeen({
                                        file: lastSeen.file,
                                        width: size.width,
                                    })
                                    return
                                }
                                const ls =
                                    files[
                                        Math.floor(
                                            (scrollTop / rowHeight) * numCols
                                        )
                                    ]
                                setLastSeen({ file: ls, width: size.width })
                            }}
                            onItemsRendered={() => {
                                if (didScroll) {
                                    return
                                }
                                // Grid ref is not ready yet even when this callback is called,
                                // but putting it in a timeout will push it off to the next tick,
                                // and the ref will be ready.
                                setTimeout(() => {
                                    if (gridRef.current && jumpTo) {
                                        const child = useFileBrowserStore
                                            .getState()
                                            .filesMap.get(jumpTo)
                                        if (child) {
                                            gridRef.current.scrollToItem({
                                                align: 'smart',
                                                rowIndex: Math.floor(
                                                    child.GetIndex() / numCols
                                                ),
                                            })
                                            setDidScroll(true)
                                        } else {
                                            console.error(
                                                'Could not find child to scroll to',
                                                jumpTo
                                            )
                                        }
                                    }
                                }, 1)
                            }}
                        >
                            {SquareWrapper}
                        </Grid>
                    )}
                </div>
            )}
        </div>
    )
}

export default FileGrid
