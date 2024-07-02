import { useEffect, useMemo, useRef, useState } from 'react';
import { FixedSizeGrid as WindowGrid } from 'react-window';
import '../components/style.scss';
import { FileDisplay } from './FileDisplay';
import { useResize } from '../components/hooks';
import { FileContextT, GlobalContextType } from './File';
import { FBDispatchT } from '../types/Types';
import './filesStyle.scss';

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
    columnIndex: number;
    rowIndex: number;
    style;
}) {
    const { thisData, index } = useMemo(() => {
        const index = rowIndex * data.globalContext.numCols + columnIndex;
        if (index < data.items.length) {
            return { thisData: data.items[index], index: index };
        } else {
            return { thisData: null, index };
        }
    }, [data, rowIndex, columnIndex]);

    if (thisData == null) {
        return null;
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
    );
}

function FileScroller({
    itemsContext,
    globalContext,
    dispatch,
}: {
    itemsContext: FileContextT[];
    globalContext: GlobalContextType;
    dispatch?: FBDispatchT;
}) {
    const windowRef = useRef(null);

    useEffect(() => {
        if (windowRef?.current && globalContext.initialScrollIndex && globalContext.numCols) {
            windowRef.current.scrollToItem({
                align: 'smart',
                rowIndex: Math.floor(globalContext.initialScrollIndex / globalContext.numCols),
            });
        }
    }, [windowRef, globalContext.initialScrollIndex, globalContext.numCols]);

    useEffect(() => {
        if (dispatch) {
            dispatch({ type: 'set_col_count', numCols: globalContext.numCols });
        }
    }, [dispatch, globalContext.numCols]);

    return (
        <div className="overflow-y-scroll no-scrollbar h-full">
            <div className="files-grid">
                {itemsContext.map((item, i) => (
                    <FileDisplay
                        key={item.file.Id()}
                        file={item.file}
                        selected={item.selected}
                        index={0}
                        context={globalContext}
                    />
                ))}
            </div>
        </div>

        // <div className="w-full h-full">
        //     <WindowGrid
        //         className="no-scrollbar"
        //         ref={windowRef}
        //         height={viewHeight}
        //         width={viewWidth}
        //         columnCount={globalContext.numCols}
        //         columnWidth={globalContext.itemWidth}
        //         rowCount={Math.ceil(itemsContext.length / globalContext.numCols)}
        //         rowHeight={globalContext.itemWidth * 1.1}
        //         itemData={{ items: itemsContext, globalContext: globalContext }}
        //         overscanRowCount={10}
        //     >
        //         {FileCell}
        //     </WindowGrid>
        // </div>
    );
}

export default FileScroller;
