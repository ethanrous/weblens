import { useEffect, useMemo, useRef, useState } from 'react'
import { FixedSizeGrid as WindowGrid } from 'react-window'
import './style.scss'
import { FileDisplay, GlobalContextType } from '../Files/FileDisplay'
import { useResize } from './hooks'
import { FileContextT } from '../Files/File'
import { FBDispatchT } from '../types/Types'

type ScrollerDataT = {
    items: FileContextT[]
    globalContext: GlobalContextType
}

function FileCell({
    data,
    columnIndex,
    rowIndex,
    style,
}: {
    data: ScrollerDataT
    columnIndex: number
    rowIndex: number
    style
}) {
    const { thisData, index } = useMemo(() => {
        const index = rowIndex * data.globalContext.numCols + columnIndex
        if (index < data.items.length) {
            return { thisData: data.items[index], index: index }
        } else {
            return { thisData: null, index }
        }
    }, [data, rowIndex, columnIndex])

    if (thisData == null) {
        return null
    }

    return (
        <div key={thisData.file.Id()} className="file-cell" style={style}>
            <FileDisplay
                key={thisData.file.Id()}
                file={thisData.file}
                selected={thisData.selected}
                index={index}
                context={{ ...data.globalContext }}
            />
        </div>
    )
}

const FILE_BASE_SIZE = 250

function FileScroller({
    itemsContext,
    globalContext,
    parentNode,
    dispatch,
}: {
    itemsContext: FileContextT[]
    globalContext: GlobalContextType
    parentNode: HTMLDivElement
    dispatch?: FBDispatchT
}) {
    const [viewWidth, setViewWidth] = useState(0)
    const [viewHeight, setViewHeight] = useState(0)
    const windowRef = useRef(null)
    const parentSize = useResize(parentNode)

    globalContext = useMemo(() => {
        const viewWidth = parentSize?.width ? parentSize.width : 0
        setViewWidth(viewWidth)
        let numCols = Math.floor((viewWidth - 1) / FILE_BASE_SIZE)
        numCols = numCols ? numCols : 1
        setViewHeight(parentSize.height ? parentSize.height : 0)
        const itemWidth = Math.floor(viewWidth / numCols)

        globalContext.numCols = numCols
        globalContext.itemWidth = itemWidth

        return globalContext
    }, [parentSize.width, parentSize.height, globalContext])

    useEffect(() => {
        if (
            windowRef?.current &&
            globalContext.initialScrollIndex &&
            globalContext.numCols
        ) {
            windowRef.current.scrollToItem({
                align: 'smart',
                rowIndex: Math.floor(
                    globalContext.initialScrollIndex / globalContext.numCols
                ),
            })
        }
    }, [windowRef, globalContext.initialScrollIndex, globalContext.numCols])

    useEffect(() => {
        if (dispatch) {
            dispatch({ type: 'set_col_count', numCols: globalContext.numCols })
        }
    }, [dispatch, globalContext.numCols])

    return (
        <div className="w-full h-full">
            <WindowGrid
                className="no-scrollbar"
                ref={windowRef}
                height={viewHeight}
                width={viewWidth}
                columnCount={globalContext.numCols}
                columnWidth={globalContext.itemWidth}
                rowCount={Math.ceil(
                    itemsContext.length / globalContext.numCols
                )}
                rowHeight={globalContext.itemWidth * 1.1}
                itemData={{ items: itemsContext, globalContext: globalContext }}
                overscanRowCount={10}
            >
                {FileCell}
            </WindowGrid>
        </div>
    )
}

export default FileScroller
