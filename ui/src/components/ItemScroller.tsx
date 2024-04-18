import { useEffect, useMemo, useRef, useState } from "react";
import { Box } from "@mantine/core";
import { FixedSizeGrid as WindowGrid } from "react-window";
import "./style.css";
import { GlobalContextType, ItemDisplay, ItemProps } from "./ItemDisplay";
import { useResize } from "./hooks";

type ItemsContextType = {
    items: ItemProps[];
    globalContext: GlobalContextType;
};

function FileCell({
    data,
    columnIndex,
    rowIndex,
    style,
}: {
    data: ItemsContextType;
    columnIndex;
    rowIndex;
    style;
}) {
    let thisData: ItemProps;
    const index = rowIndex * data.globalContext.numCols + columnIndex;
    if (index < data.items.length) {
        thisData = data.items[index];
        thisData.index = index;
    } else {
        return null;
    }

    return (
        <Box key={thisData.itemId} className="file-cell" style={style}>
            <ItemDisplay
                key={thisData.itemId}
                itemInfo={thisData}
                context={{ ...data.globalContext }}
            />
        </Box>
    );
}

const FILE_BASE_SIZE = 250;

export const ItemScroller = ({
    itemsContext,
    globalContext,
    parentNode,
    dispatch,
}: {
    itemsContext: ItemProps[];
    globalContext: GlobalContextType;
    parentNode;
    dispatch?;
}) => {
    const [viewWidth, setViewWidth] = useState(0);
    const [viewHeight, setViewHeight] = useState(0);
    const windowRef = useRef(null);
    const parentSize = useResize(parentNode);

    globalContext = useMemo(() => {
        const viewWidth = parentSize?.width ? parentSize.width : 0;
        setViewWidth(viewWidth);
        var numCols = Math.floor((viewWidth - 1) / FILE_BASE_SIZE);
        numCols = numCols ? numCols : 1;
        setViewHeight(parentSize.height ? parentSize.height : 0);
        const itemWidth = Math.floor(viewWidth / numCols);

        globalContext.numCols = numCols;
        globalContext.itemWidth = itemWidth;

        return globalContext;
    }, [parentSize.width, parentSize.height, globalContext]);

    useEffect(() => {
        if (
            windowRef?.current &&
            globalContext.initialScrollIndex &&
            globalContext.numCols
        ) {
            windowRef.current.scrollToItem({
                align: "smart",
                rowIndex: Math.floor(
                    globalContext.initialScrollIndex / globalContext.numCols
                ),
            });
        }
    }, [windowRef, globalContext.initialScrollIndex, globalContext.numCols]);

    useEffect(() => {
        if (dispatch) {
            dispatch({ type: "set_col_count", numCols: globalContext.numCols });
        }
    }, [dispatch, globalContext.numCols]);

    return (
        <Box style={{ width: "100%", height: "100%" }}>
            <WindowGrid
                className="no-scrollbars"
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
        </Box>
    );
};
