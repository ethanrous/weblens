import { useEffect, useMemo, useState, memo } from 'react'

import FolderIcon from '@mui/icons-material/Folder'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import { styled, Tooltip, FormControl, Typography, Input, Skeleton, Box, Divider } from '@mui/joy'

import { StyledLazyThumb } from '../../types/Styles'
import { humanFileSize } from '../../util'
import { MoveFile, RenameFile } from '../../api/FileBrowserApi'
import { FileItemWrapper, ItemVisualComponentWrapper } from './FilebrowserStyles'
import { itemData } from '../../types/Types'

import { useSnackbar } from 'notistack'

const StyledInputBase = styled(Input)(({ theme }) => ({
    '& .MuiInputBase-input': {
        color: theme.palette.primary.mainChannel,
    },
}))

function StartKeybaordListener(dispatch, editing, newName, filepath, authHeader) {

    const keyDownHandler = event => {
        if (!editing) { return }
        if (event.key === 'Enter') {
            event.preventDefault()
            if (newName == "") {
                dispatch({ type: 'reject_edit' })
            } else {
                let newPath = RenameFile(filepath, newName, authHeader)
                dispatch({ type: 'confirm_edit', newPath: newPath, file: filepath })
            }
        }
    }

    window.addEventListener('keydown', keyDownHandler)

    return () => { window.removeEventListener('keydown', keyDownHandler) }
}

const ItemVisualComponent = ({ itemData, type, isDir, imported }) => {
    if (isDir) {
        return (<FolderIcon style={{ width: "80%", height: "80%" }} />)
    } else if (type == "File") {
        return (<InsertDriveFileIcon style={{ width: "80%", height: "80%" }} />)
    } else if (imported) {
        return (<StyledLazyThumb mediaData={itemData.mediaData} quality={"thumbnail"} lazy={false} />)
    } else {
        return (<Skeleton animation="wave" height={"100%"} width={"100%"} variant="rectangular" />)
    }
}

const EditingHook = ({ dispatch }) => {
    let focused = false
    const [previous, setPrevious] = useState(false)

    useEffect(() => {
        if (!focused && previous == true) {
            dispatch({ type: 'reject_edit' })
        } else {
            setPrevious(focused)
        }
    }, [focused])
    return null
}

const TextBox = ({ itemData, editing, setRenameVal, dispatch }) => {
    const filename = useMemo(() => {
        return itemData.filepath.substring(itemData.filepath.lastIndexOf('/') + 1)
    }, [itemData.filepath])

    if (editing) {
        return (
            <FormControl style={{ width: "90%" }}>
                <StyledInputBase
                    autoFocus={true}
                    defaultValue={filename}
                    onClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}
                />
                <EditingHook dispatch={dispatch} />
            </FormControl>
        )
    } else {
        const [sizeValue, units] = humanFileSize(itemData.size, true)
        return (
            <Tooltip title={filename} disableInteractive >
                <Box
                    display={"flex"}
                    flexDirection={"column"}
                    justifyContent={"center"}
                    alignItems={"center"}

                    width={"100%"}
                    onClick={(e) => { e.stopPropagation(); dispatch({ type: 'start_editing', file: itemData.filepath }) }}
                    sx={{ cursor: 'text' }}
                >
                    <Box display={"flex"} justifyContent={"space-evenly"} alignItems={'center'} width={"100%"}>
                        <Typography fontSize={15} noWrap sx={{ color: "white" }}>{filename} </Typography>
                        <Divider orientation='vertical' sx={{ marginLeft: '6px', marginRight: '6px' }} />
                        <Box display={"flex"} flexDirection={'column'} alignContent={'center'} alignItems={'center'} >
                            <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible' }}> {sizeValue} </Typography>
                            <Typography fontSize={10} noWrap sx={{ color: "white", overflow: 'visible' }}> {units} </Typography>
                        </Box>
                    </Box>
                </Box>
            </Tooltip>
        )
    }
}

const Item = memo(({ itemData, selected, editing, dragging, dispatch, authHeader }: { itemData: itemData, selected: boolean, editing: boolean, dragging: number, dispatch: any, authHeader: any }) => {
    const [hovering, setHovering] = useState(false)
    const [mouseDown, setMouseDown] = useState(false)
    const [renameVal, setRenameVal] = useState("")
    const { enqueueSnackbar } = useSnackbar()

    useEffect(() => {
        if (editing) {
            return StartKeybaordListener(dispatch, editing, renameVal, itemData.filepath, authHeader)
        }
    }, [editing, renameVal])

    useEffect(() => {
        if (itemData.updatePath && itemData.filepath != itemData.updatePath) {
            MoveFile(itemData.filepath, itemData.updatePath, authHeader).then((res) => {
                if (res.status != 200) {
                    return Promise.reject(`Could not move file: ${res.statusText}`)
                }
                dispatch({ type: "delete_from_map", item: itemData.filepath })
                itemData.filepath = itemData.updatePath
                itemData.updatePath = ''
            }).catch(r => { itemData.updatePath = ''; enqueueSnackbar(`${r} (${itemData.filepath.slice(itemData.filepath.lastIndexOf('/') + 1)})`, { variant: "error" }) })
        }
    }, [itemData.updatePath])

    return (
        <FileItemWrapper
            hovering={hovering} isDir={itemData.isDir} selected={selected} dragging={dragging}
            onMouseDown={() => { setMouseDown(true) }}
            onMouseUp={() => { dispatch({ type: 'move_selected', targetItemPath: itemData.filepath }); setMouseDown(false) }}
            onMouseOver={(e) => { e.stopPropagation(); setHovering(true); dispatch({ type: 'set_hovering', itempath: itemData.filepath }) }}
            onClick={(e) => { e.stopPropagation(); dispatch({ type: 'set_selected', itempath: itemData.filepath }) }}
            onDoubleClick={() => { if (itemData.isDir) { dispatch({ type: 'set_path', path: itemData.filepath }) } else { dispatch({ type: 'set_presentation', presentingPath: itemData.filepath }) } }}
            onMouseLeave={(e) => {
                setHovering(false); if (!selected && mouseDown) { dispatch({ type: "clear_selected" }) }
                if (mouseDown) {
                    dispatch({ type: 'set_selected', itempath: itemData.filepath, selected: true })
                    dispatch({ type: 'set_dragging', dragging: true })
                    setMouseDown(false)
                }
            }}
        >
            <ItemVisualComponentWrapper>
                <ItemVisualComponent itemData={itemData} type={itemData.mediaData?.MediaType.FriendlyName} isDir={itemData.isDir} imported={itemData.imported} />
            </ItemVisualComponentWrapper>

            <TextBox itemData={itemData} editing={editing} setRenameVal={setRenameVal} dispatch={dispatch} />
        </FileItemWrapper>
    )
}, (prev, next) => {
    if (prev.itemData.updatePath != next.itemData.updatePath) {
        return false
    } else if (prev.selected != next.selected) {
        return false
    } else if (prev.editing != next.editing) {
        return false
    } else if (prev.dragging != next.dragging) {
        return false
    } else if (prev.itemData.imported != next.itemData.imported) {
        return false
    }
    return true
})

export default Item