import { useEffect, useMemo, useState, memo, useRef } from 'react'

import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import { FormControl, Typography, Input, Box, Divider } from '@mui/joy'

import { StyledLazyThumb } from '../../types/Styles'
import { humanFileSize } from '../../util'
import { CreateFolder, RenameFile } from '../../api/FileBrowserApi'
import { FileItemWrapper, ItemVisualComponentWrapper } from './FilebrowserStyles'
import { itemData } from '../../types/Types'

import { useSnackbar } from 'notistack'
import { useIsVisible } from '../../components/PhotoContainer'
import { IconFileZip, IconFolder } from '@tabler/icons-react'
import { Center, Skeleton, Text } from '@mantine/core'

function StartKeybaordListener(dispatch, editing, parentId, itemId, oldName, newName, imported, authHeader) {

    const keyDownHandler = event => {
        if (!editing) { return }
        if (event.key === 'Enter') {
            event.preventDefault()
            if (newName === "") {
                dispatch({ type: 'reject_edit' })
            } else {
                if (imported) {
                    RenameFile(parentId, oldName, newName, authHeader).then(newId => dispatch({ type: 'confirm_edit', itemId: itemId, newItemId: newId, newFilename: newName }))
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

const ItemVisualComponent = ({ itemData }: { itemData: itemData }) => {
    const sqareSize = "75%"
    const type = itemData.mediaData?.MediaType.FriendlyName
    const displayable = itemData.mediaData?.MediaType.IsDisplayable
    if (itemData.isDir) {
        return (<IconFolder style={{ width: sqareSize, height: sqareSize }} />)
    } else if (displayable) {
        return (<StyledLazyThumb mediaData={itemData.mediaData} quality={"thumbnail"} lazy={true} />)
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

const TextBox = ({ itemData, editing, setRenameVal, dispatch }: { itemData: itemData, editing: boolean, setRenameVal: any, dispatch: any }) => {
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
                    defaultValue={itemData.filename}
                    onClick={(e) => { e.stopPropagation() }}
                    onDoubleClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}
                />
                <EditingHook dispatch={dispatch} />
            </FormControl>
        )
    } else {
        const [sizeValue, units] = humanFileSize(itemData.size, true)
        return (
            <Box
                display={"flex"}
                flexDirection={"column"}
                justifyContent={"center"}
                alignItems={"center"}

                width={"100%"}
                onClick={(e) => { e.stopPropagation(); dispatch({ type: 'start_editing', fileId: itemData.id }) }}
                sx={{ cursor: 'text' }}
            >
                <Box display={"flex"} justifyContent={"space-evenly"} alignItems={'center'} width={"100%"} height={"30px"}>
                    <Typography fontSize={15} noWrap sx={{ color: "white", userSelect: 'none' }}>{itemData.filename} </Typography>
                    <Divider orientation='vertical' sx={{ marginLeft: '6px', marginRight: '6px' }} />
                    <Box display={"flex"} flexDirection={'column'} alignContent={'center'} alignItems={'center'} >
                        <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {sizeValue} </Typography>
                        <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {units} </Typography>
                    </Box>
                </Box>
            </Box>
        )
    }
}

const Item = memo(({ itemData, selected, moveSelected, editing, dragging, dispatch, authHeader }: { itemData: itemData, selected: boolean, moveSelected: () => void, editing: boolean, dragging: number, dispatch: any, authHeader: any }) => {
    const [hovering, setHovering] = useState(false)
    const [renameVal, setRenameVal] = useState("")
    const { enqueueSnackbar } = useSnackbar()
    const itemRef = useRef()
    const isVisible = useIsVisible(itemRef, false)

    useEffect(() => {
        dispatch({ type: "set_visible", item: itemData.id, visible: isVisible })
    }, [isVisible])

    useEffect(() => {
        if (editing) {
            return StartKeybaordListener(dispatch, editing, itemData.parentFolderId, itemData.id, itemData.filename, renameVal, itemData.imported, authHeader)
        }
    }, [editing, renameVal])

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
                <ItemVisualComponent itemData={itemData} />
            </ItemVisualComponentWrapper>

            <TextBox itemData={itemData} editing={editing} setRenameVal={setRenameVal} dispatch={dispatch} />
        </FileItemWrapper>
    )
}, (prev, next) => {
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
    }
    return true
})

export default Item