import { useEffect, useMemo, useRef, useState } from "react";
import { Box } from "@mantine/core";
import { FixedSizeGrid as WindowGrid } from "react-window";
import "./style.css";
import { GlobalContextType, FileDisplay, SelectedState } from "./ItemDisplay";
import { useResize } from "./hooks";
import { WeblensFile } from "../classes/File";

export type FileContextT = {
    file: WeblensFile;
    selected: SelectedState;
};

type ScrollerDataT = {
    items: FileContextT[];
    globalContext: GlobalContextType;
};

function FileCell({
    data,
    columnIndex,
    rowIndex,
    style,
}: {
    data: ScrollerDataT;
    columnIndex;
    rowIndex;
    style;
}) {
    let thisData: FileContextT;
    const index = rowIndex * data.globalContext.numCols + columnIndex;
    if (index < data.items.length) {
        thisData = data.items[index];
    } else {
        return null;
    }

    return (
        <Box key={thisData.file.Id()} className="file-cell" style={style}>
            <FileDisplay
                key={thisData.file.Id()}
                file={thisData.file}
                selected={thisData.selected}
                index={index}
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
    itemsContext: FileContextT[];
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
