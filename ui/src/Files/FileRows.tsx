import { Box, Space } from '@mantine/core'
import { FixedSizeList as ScrollList } from 'react-window'
import { FBDispatchT, FbStateT } from '../types/Types'
import { WeblensFile } from './File'
import { useContext, useEffect, useMemo, useState } from 'react'
import { IconDisplay } from '../Pages/FileBrowser/FileBrowserStyles'
import { useNavigate } from 'react-router-dom'
import { useResize } from '../components/hooks'
import { UserContext } from '../Context'
import { IconTrash } from '@tabler/icons-react'

function FileRow({
    data,
    index,
    style,
}: {
    data: {
        files: WeblensFile[]
        selected: Map<string, boolean>
        lastSelectedIndex: number
        hoveringIndex: number
        dispatch: FBDispatchT
    }
    index: number
    style
}) {
    const nav = useNavigate()
    const { usr } = useContext(UserContext)
    const inRange =
        (index - data.lastSelectedIndex) * (index - data.hoveringIndex) < 1

    return (
        <div style={style} className="flex">
            <div
                className="file-row"
                data-selected={
                    data.selected.get(data.files[index].Id()) === true
                }
                data-in-range={inRange}
                style={{
                    border:
                        index === data.lastSelectedIndex
                            ? '2px solid #442299'
                            : '2px solid #00000000',
                }}
                onClick={(e) => {
                    e.stopPropagation()
                    data.dispatch({
                        type: 'set_selected',
                        fileId: data.files[index].Id(),
                        selected:
                            data.selected.get(data.files[index].Id()) !== true,
                    })
                }}
                onMouseOver={(e) => {
                    data.dispatch({
                        type: 'set_hovering',
                        hovering: data.files[index].Id(),
                    })
                }}
                onContextMenu={(e) => {
                    e.stopPropagation()
                    e.preventDefault()
                    data.dispatch({
                        type: 'set_menu_target',
                        fileId: data.files[index].Id(),
                    })
                    data.dispatch({ type: 'set_menu_open', open: true })
                    data.dispatch({
                        type: 'set_menu_pos',
                        pos: { x: e.clientX, y: e.clientY },
                    })
                }}
                onDoubleClick={() => {
                    if (data.files[index].IsFolder()) {
                        nav(`/files/${data.files[index].Id()}`)
                    } else {
                        nav(
                            `/files/${data.files[
                                index
                            ].ParentId()}?jumpTo=${data.files[index].Id()}`
                        )
                    }
                }}
            >
                <div className="w-6 h-6">
                    <IconDisplay file={data.files[index]} />
                </div>
                <p>{data.files[index].GetFilename()}</p>
                <Space flex={1} />
                {data.files[index].ParentId() === usr.trashId && <IconTrash />}
                <p>{data.files[index].FormatSize()}</p>
            </div>
        </div>
    )
}

export function FileRows({
    fb,
    dispatch,
}: {
    fb: FbStateT
    dispatch: FBDispatchT
}) {
    const [boxRef, setBoxRef] = useState(null)
    const size = useResize(boxRef)

    const files = useMemo(() => {
        return Array.from(fb.dirMap.values())
    }, [fb.dirMap])

    useEffect(() => {
        dispatch({
            type: 'set_files_list',
            fileIds: files.map((v) => v.Id()),
        })
    }, [files])

    const lastSelectedIndex = useMemo(() => {
        return files.findIndex((v) => v.Id() === fb.lastSelected)
    }, [files, fb.lastSelected])

    const hoveringIndex = useMemo(() => {
        if (!fb.holdingShift || !files) {
            return { hoveringIndex: -1, lastSelectedIndex: -1 }
        }
        return files.findIndex((v) => v.Id() === fb.hovering)
    }, [files, fb.holdingShift, fb.hovering])

    return (
        <Box
            ref={setBoxRef}
            className="flex flex-col items-center w-full h-full ml-1"
        >
            <ScrollList
                className="no-scrollbar"
                height={boxRef?.clientHeight ? boxRef.clientHeight : 0}
                width={size.width}
                itemSize={52}
                itemCount={files.length}
                itemData={{
                    files: files,
                    selected: fb.selected,
                    lastSelectedIndex: lastSelectedIndex,
                    hoveringIndex: hoveringIndex,
                    dispatch: dispatch,
                }}
                overscanRowCount={50}
            >
                {FileRow}
            </ScrollList>
        </Box>
    )
}
