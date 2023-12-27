import { useEffect, useState, memo, useRef } from 'react'

import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import { FormControl, Typography, Input, Divider } from '@mui/joy'

import { StyledLazyThumb } from '../../types/Styles'
import { humanFileSize } from '../../util'
import { CreateFolder, RenameFile } from '../../api/FileBrowserApi'
import { FileItemWrapper, FlexColumnBox, ItemVisualComponentWrapper } from './FilebrowserStyles'
import { itemData } from '../../types/Types'

import { useIsVisible } from '../../components/PhotoContainer'
import { IconFileZip, IconFolder } from '@tabler/icons-react'
import { Box, Center, Skeleton, Text, Tooltip } from '@mantine/core'

function StartKeybaordListener(dispatch, editing, parentId, itemId, oldName, newName, imported, authHeader) {

    const keyDownHandler = event => {
        if (!editing) { return }
        if (event.key === 'Enter') {
            event.preventDefault()
            if (newName === "") {
                dispatch({ type: 'reject_edit' })
            } else {
                if (imported) {
                    RenameFile(parentId, oldName, newName, authHeader)
                } else {
                    CreateFolder(parentId, newName, authHeader).then(res => {
                        dispatch({ type: 'set_selected', itemId: res.folderId })
                        dispatch({ type: 'confirm_edit', itemId: itemId })
                    })
                }
            }
        }
    }

    window.addEventListener('keydown', keyDownHandler)

    return () => { window.removeEventListener('keydown', keyDownHandler) }
}

const ItemVisualComponent = ({ itemData, root }: { itemData: itemData, root }) => {
    const sqareSize = "75%"
    const type = itemData.mediaData?.mediaType.FriendlyName
    const displayable = itemData.mediaData?.mediaType.IsDisplayable
    if (itemData.isDir) {
        return (<IconFolder style={{ width: sqareSize, height: sqareSize }} />)
    } else if (displayable) {
        return (<StyledLazyThumb mediaData={itemData.mediaData} quality={"thumbnail"} lazy={true} root={root} />)
    } else if (type === "File") {
        return (<InsertDriveFileIcon style={{ width: sqareSize, height: sqareSize }} />)
    } else if (type === "Zip") {
        return (<IconFileZip style={{ width: sqareSize, height: sqareSize }} />)
    } else {
        return (
            <Center style={{ height: "100%", width: "100%" }}>
                <Skeleton height={"100%"} width={"100%"} />
                <Text pos={'absolute'} style={{ userSelect: 'none' }}>Processing...</Text>
            </Center>
        )
    }
}

const EditingHook = ({ dispatch }) => {
    let focused = false
    const [previous, setPrevious] = useState(false)

    useEffect(() => {
        if (!focused && previous === true) {
            dispatch({ type: 'reject_edit' })
        } else {
            setPrevious(focused)
        }
    }, [focused])
    return null
}

const TextBox = ({ filename, fileId, fileSize, editing, setRenameVal, dispatch }: { filename: string, fileId: string, fileSize: number, editing: boolean, setRenameVal: any, dispatch: any }) => {
    const editRef: React.Ref<HTMLInputElement> = useRef()
    useEffect(() => {
        if (editRef.current) {
            editRef.current.select()
        }
    }, [editing])

    if (editing) {
        return (
            <FormControl style={{ width: "100%", height: "30px", bottom: "5px" }} >
                <Input
                    slotProps={{ input: { ref: editRef } }}
                    autoFocus={true}
                    defaultValue={filename}
                    onClick={(e) => { e.stopPropagation() }}
                    onDoubleClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}
                />
                <EditingHook dispatch={dispatch} />
            </FormControl>
        )
    } else {
        const [sizeValue, units] = humanFileSize(fileSize, true)
        return (
            <FlexColumnBox style={{ width: '100%', cursor: 'text', padding: '5px' }}>
                <Tooltip openDelay={300} label={filename}>
                    <Box display={"flex"} style={{ justifyContent: 'space-evenly', alignItems: 'center', width: '100%', height: '30px' }} onClick={(e) => { e.stopPropagation(); dispatch({ type: 'start_editing', fileId: fileId }) }}>
                        <Typography fontSize={15} noWrap sx={{ color: "white", userSelect: 'none' }}>{filename} </Typography>
                        <Divider orientation='vertical' sx={{ marginLeft: '6px', marginRight: '6px' }} />
                        <FlexColumnBox >
                            <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {sizeValue} </Typography>
                            <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {units} </Typography>
                        </FlexColumnBox>
                    </Box>
            </Tooltip>
            </FlexColumnBox>
        )
    }
}

const Item = memo(({ itemData, selected, root, moveSelected, editing, dragging, dispatch, authHeader, visual }: {
    itemData: itemData, selected: boolean, root, moveSelected: () => void, editing: boolean, dragging: number, dispatch: any, authHeader: any, visual?: JSX.Element
}) => {
    const [hovering, setHovering] = useState(false)
    const [renameVal, setRenameVal] = useState("")
    const itemRef = useRef()
    const { isVisible } = useIsVisible(root, itemRef, false)

    useEffect(() => {
        dispatch({ type: "set_visible", item: itemData.id, visible: isVisible })
    }, [isVisible])

    useEffect(() => {
        if (editing) {
            return StartKeybaordListener(dispatch, editing, itemData.parentFolderId, itemData.id, itemData.filename, renameVal, itemData.imported, authHeader)
        }
    }, [editing, renameVal])

    if (!visual) {
        visual = <ItemVisualComponent itemData={itemData} root={root} />
    }

    return (
        <FileItemWrapper
            itemRef={itemRef}
            itemData={itemData}
            dispatch={dispatch}
            hovering={hovering}
            setHovering={setHovering}
            isDir={itemData.isDir}
            selected={selected}
            moveSelected={moveSelected}
            dragging={dragging}
        >
            <ItemVisualComponentWrapper>
                {visual}
            </ItemVisualComponentWrapper>

            <TextBox filename={itemData.filename} fileId={itemData.id} fileSize={itemData.size} editing={editing} setRenameVal={setRenameVal} dispatch={dispatch} />
        </FileItemWrapper>
    )
}, (prev, next) => {
    if (prev.itemData.visible !== next.itemData.visible) {
        return false
    }

    if (!next.itemData.visible) {
        return true
    }

    if (prev.selected !== next.selected) {
        return false
    } else if (prev.editing !== next.editing) {
        return false
    } else if (prev.dragging !== next.dragging) {
        return false
    } else if (prev.itemData.imported !== next.itemData.imported) {
        return false
    } else if (prev.itemData.size !== next.itemData.size) {
        return false
    }
    return true
})

export default Item