import { WeblensFile } from '@weblens/types/files/File'
import 'components/style.scss'
import '@weblens/types/files/filesStyle.scss'
import { FileSquare } from '@weblens/types/files/FileSquare'
import { useResize } from 'components/hooks'
import { useEffect, useMemo, useRef, useState } from 'react'
import { FixedSizeGrid as Grid } from 'react-window'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { GetStartedCard } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'

function SquareWrapper({ data, rowIndex, columnIndex, style }) {
    if (!data || rowIndex === undefined) {
        return null
    }

    const absIndex = rowIndex * data.numCols + columnIndex
    if (absIndex > data.files.length - 1) {
        return null
    }
    const file = data.files[absIndex]
    if (!file) {
        console.error('Cant find grid file at', rowIndex, columnIndex)
        return null
    }

    return (
        <div style={style}>
            <FileSquare file={file} />
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

    if (files.length === 0) {
        return <GetStartedCard />
    }

    return (
        <div ref={setContainerRef} className="h-full w-full outline-0">
            {size.width !== -1 && (
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
                            files[Math.floor((scrollTop / rowHeight) * numCols)]
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
    )
}

export default FileGrid
