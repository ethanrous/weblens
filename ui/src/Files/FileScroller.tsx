import { FileSquare } from './FileSquare'
import { WeblensFile } from './File'
import '../components/style.scss'
import './filesStyle.scss'

function FileScroller({ files }: { files: WeblensFile[] }) {
    // TODO - scroll to index
    // useEffect(() => {
    //     if (windowRef?.current && globalContext.initialScrollIndex && globalContext.numCols) {
    //         windowRef.current.scrollToItem({
    //             align: 'smart',
    //             rowIndex: Math.floor(globalContext.initialScrollIndex / globalContext.numCols),
    //         });
    //     }
    // }, [windowRef, globalContext.initialScrollIndex, globalContext.numCols]);

    // useEffect(() => {
    //     if (dispatch) {
    //         dispatch({ type: 'set_col_count', numCols: globalContext.numCols });
    //     }
    // }, [dispatch, globalContext.numCols]);

    return (
        <div className="overflow-y-scroll no-scrollbar h-full">
            <div className="files-grid">
                {files.map((file) => (
                    <FileSquare key={file.Id()} file={file} />
                ))}
            </div>
        </div>
    )
}

export default FileScroller
