import { WeblensFile } from './File'
import '../components/style.scss'
import './filesStyle.scss'
import { FixedSizeGrid as Grid, List } from 'react-window'
import { useFileBrowserStore } from '../Pages/FileBrowser/FBStateControl'
import { useEffect, useState } from 'react'
import { useResize } from '../components/hooks'
import { FileSquare } from './FileSquare'

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
    // TODO - scroll to index

    // const numCols = useFileBrowserStore((state) => state.numCols)
    const scrollTo = useFileBrowserStore((state) => state.scrollTo)

    const setNumCols = useFileBrowserStore((state) => state.setNumCols)

    const [gridRef, setGridRef] = useState<List>()
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const size = useResize(containerRef)

    const numCols = Math.max(Math.floor(size.width / 250), 2)

    useEffect(() => {
        // if (gridRef && numCols && scrollTo) {
        //     const scrollIndex = files.findIndex((f) => f.Id() === scrollTo)
        //     gridRef.scrollToIndex(scrollIndex)
        // }
    }, [gridRef, numCols, scrollTo])

    // useEffect(() => {
    //     setNumCols()
    //     if (dispatch) {
    //         dispatch({ type: 'set_col_count', numCols: globalContext.numCols })
    //     }
    // }, [dispatch, globalContext.numCols])
    useEffect(() => {
        setNumCols(4)
    }, [])

    const squareSize = (size.width / numCols) * 1.15
    const margin = 8

    return (
        <div ref={setContainerRef} className="h-full w-full outline-0">
            {size.width !== -1 && (
                <Grid
                    className="no-scrollbar outline-0"
                    ref={setGridRef}
                    columnCount={numCols}
                    itemData={{ files, numCols }}
                    height={size.height}
                    width={size.width}
                    rowCount={Math.ceil(files.length / numCols)}
                    columnWidth={size.width / numCols}
                    rowHeight={squareSize + margin}
                >
                    {SquareWrapper}
                </Grid>
            )}
        </div>
    )
}

export default FileGrid