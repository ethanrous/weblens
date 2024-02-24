import { useEffect, useMemo, useRef, useState } from 'react'
import { Box } from '@mantine/core'
import { FixedSizeGrid as WindowGrid } from "react-window"
import './style.css'
import { GlobalContextType, ItemDisplay, ItemProps } from './ItemDisplay'

type ItemsContextType = {
    items: ItemProps[]
    globalContext: GlobalContextType
}

export const useWindowSize = (setWindowSize) => {
    useEffect(() => {
        window.addEventListener('resize', setWindowSize);
        return () => window.removeEventListener('resize', setWindowSize);
    }, [setWindowSize]);
}

function FileCell({ data, columnIndex, rowIndex, style }: { data: ItemsContextType, columnIndex, rowIndex, style }) {
    let thisData
    const index = (rowIndex * data.globalContext.numCols) + columnIndex
    if (index < data.items.length) {
        thisData = data.items[index]
    } else {
        return null
    }

    return (
        <Box className="file-cell" style={style}>
            <ItemDisplay itemInfo={thisData} context={{ ...data.globalContext }} />
        </Box>
    )
}

const FILE_BASE_SIZE = 250

export const ItemScroller = ({ itemsContext, globalContext, dispatch }: { itemsContext: ItemProps[], globalContext: GlobalContextType, dispatch?}) => {
    const [boxNode, setBoxNode] = useState(null)
    const [, setWindowSize] = useState(null)

    const [viewWidth, setViewWidth] = useState(0)
    const [viewHeight, setViewHeight] = useState(0)
    const windowRef = useRef(null)
    useWindowSize(setWindowSize)

    globalContext = useMemo(() => {
        const viewWidth = boxNode?.clientWidth ? boxNode.clientWidth : 0
        setViewWidth(viewWidth)
        var numCols = Math.floor((viewWidth - 1) / FILE_BASE_SIZE)
        numCols = numCols ? numCols : 1
        setViewHeight(boxNode?.clientHeight ? boxNode.clientHeight : 0)
        const itemWidth = Math.floor(viewWidth / numCols)

        globalContext.numCols = numCols
        globalContext.itemWidth = itemWidth

        return globalContext
    }, [boxNode?.clientWidth, boxNode?.clientHeight, globalContext.dragging])

    useEffect(() => {
        if (windowRef?.current && globalContext.initialScrollIndex && globalContext.numCols) {
            windowRef.current.scrollToItem({ align: "smart", rowIndex: Math.floor(globalContext.initialScrollIndex / globalContext.numCols) })
        }
    }, [windowRef, globalContext.initialScrollIndex, globalContext.numCols])

    useEffect(() => {
        if (dispatch) {
            dispatch({ type: 'set_col_count', numCols: globalContext.numCols })
        }
    }, [dispatch, globalContext.numCols])

    return (
        <Box ref={setBoxNode} style={{ width: '100%', height: '100%' }}>
            <WindowGrid
                className="no-scrollbars"

                ref={windowRef}
                height={viewHeight - 58}
                width={viewWidth}

                columnCount={globalContext.numCols}
                columnWidth={globalContext.itemWidth}

                rowCount={Math.ceil(itemsContext.length / globalContext.numCols)}
                rowHeight={(globalContext.itemWidth) * 1.10}
                itemData={{ items: itemsContext, globalContext: globalContext }}

                overscanRowCount={10}
            >
                {FileCell}
            </WindowGrid>
        </Box>
    )
}
