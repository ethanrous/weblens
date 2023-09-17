import { useEffect, useMemo, useState } from 'react'

import Box from '@mui/material/Box'
import FolderIcon from '@mui/icons-material/Folder'
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile'
import Skeleton from '@mui/material/Skeleton'
import Checkbox from '@mui/material/Checkbox'
import InputBase from '@mui/material/InputBase'
import Typography from '@mui/material/Typography'
import FormControl, { useFormControl } from '@mui/material/FormControl'
import Tooltip from '@mui/material/Tooltip'

import { StyledLazyThumb } from '../../types/Styles'
import { humanFileSize, dateFromItemData } from '../../util'
// import { itemData } from '../../types/FileBrowserTypes'

const boxSX = {
    outline: "1px solid #00F0FF",
    color: 'gray',
    backgroundColor: 'lightblue'
}

// type ItemProps = {
//     itemData: itemData
//     dispatch: React.Dispatch<fileBrowserAction>
//     navigate: NavigateFunction
// }

function StartKeybaordListener(dispatch, editing, newName, filepath) {

    const keyDownHandler = event => {
        if (!editing) { return }
        switch (event.key) {
            case 'Escape': {
                event.preventDefault()
                dispatch({ type: 'reject_edit' })
                return
            }
            case 'Enter': {
                event.preventDefault()
                if (newName == "") {
                    dispatch({ type: 'reject_edit' })
                } else {
                    dispatch({ type: 'confirm_edit', newName: newName, file: filepath })
                }
                return
            }
        }

    }

    window.addEventListener('keydown', keyDownHandler)

    return () => { window.removeEventListener('keydown', keyDownHandler) }
}

const ItemVisualComponent = ({ itemData, type, isDir, imported }) => {
    if (isDir) {
        return (<FolderIcon style={{ width: "80%", height: "80%", cursor: "pointer", marginBottom: "20%" }} onDragOver={() => { }} />)
    } else if (type == "File") {
        return (<InsertDriveFileIcon style={{ width: "80%", height: "80%", cursor: "pointer", marginBottom: "20%" }} onDragOver={() => { }} />)
    } else if (imported) {
        return (<StyledLazyThumb mediaData={itemData} quality={"thumbnail"} lazy={true} dispatch={() => { }} />)//width={"200px"} height={"200px"} sx={{ cursor: "pointer" }} />)
    } else {
        return (<Skeleton animation="wave" height={"100%"} width={"100%"} variant="rectangular" />)
    }
}

const EditingHook = ({ dispatch }) => {
    const { focused } = useFormControl() || {}
    const [previous, setPrevious] = useState(false)

    useEffect(() => {
        if (!focused && previous == true) {
            dispatch({ type: 'reject_edit' })
        } else {
            setPrevious(focused)
        }
    }, [focused])
    return (<></>)
}

const TextBox = ({ itemData, editing, hasInfo, setRenameVal, dispatch }) => {

    const filename = useMemo(() => {
        return itemData.filepath.substring(itemData.filepath.lastIndexOf('/') + 1)
    }, [itemData.filepath])

    if (editing) {
        let periodIndex = filename.lastIndexOf('.')

        if (periodIndex != -1) {
            var ext = filename.slice(periodIndex)
            var basename = filename.slice(0, periodIndex)
        } else {
            var basename = filename
        }

        return (
            <FormControl style={{ width: "90%" }}>
                <InputBase
                    autoFocus
                    placeholder={basename}
                    onClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}

                    endAdornment={ext ? ext : ""}
                />
                <EditingHook dispatch={dispatch} />
            </FormControl>
        )
    } else if (hasInfo) {
        return (
            <Tooltip title={filename} enterNextDelay={100}>
                <Box
                    display={"flex"}
                    flexDirection={"column"}
                    justifyContent={"center"}
                    height={"90%"}
                    width={"90%"}
                    onClick={(e) => { e.stopPropagation(); dispatch({ type: 'start_editing', file: itemData.filepath }) }}
                >
                    <Typography noWrap style={{ margin: 0, color: "white", cursor: "text" }}>{filename} </Typography>
                    <Box display={"flex"} justifyContent={"space-evenly"}>
                        <Typography fontSize={12} noWrap style={{ margin: 0, color: "white", cursor: "text" }}>{humanFileSize(itemData.size)} </Typography>
                        <Typography fontSize={12} noWrap style={{ margin: 0, color: "white", cursor: "text" }}>{dateFromItemData(itemData)} </Typography>
                    </Box>
                </Box>
            </Tooltip>
        )
    } else {
        return (<Skeleton animation="wave" height={10} width="40%" />)
    }
}

export default function Item({ itemData, editing, dispatch, anyChecked, navigate }) {
    const hasInfo = useMemo(() => {
        return itemData.mediaData ? true : false
    }, [itemData.mediaData])

    const [hovering, setHovering] = useState(false)
    const [renameVal, setRenameVal] = useState("")

    useEffect(() => {
        if (editing) {
            return StartKeybaordListener(dispatch, editing, renameVal, itemData.filepath)
        }
    }, [editing, renameVal])

    const unselectedAction = useMemo(() => {
        let unselectedAction
        if (itemData.isDir) {
            unselectedAction = () => navigate(("/files/" + itemData.filepath).replace(/\/\/+/g, '/'))
        } else if (itemData.imported) {
            unselectedAction = () => dispatch({ type: 'set_presentation', presentingHash: itemData.mediaData.FileHash })
        } else {
            unselectedAction = () => { }
        }
        return unselectedAction
    }, [])

    const select = (e) => { dispatch({ type: 'set_selected', itempath: itemData.filepath, selected: !itemData.selected }) }

    return (
        <Box
            position={"relative"}
            display={"flex"}
            justifyContent={"center"}
            height={"200px"}
            width={"200px"}
            overflow={"hidden"}
            borderRadius={"10px"}
            margin={"10px"}
            sx={itemData.selected ? boxSX : {}}
            onMouseOver={() => setHovering(true)}
            onMouseLeave={() => setHovering(false)}
            onClick={anyChecked ? select : () => { }}
            onContextMenu={(e) => { e.preventDefault(); e.stopPropagation() }}
        >
            {(hovering || itemData.selected) && hasInfo && (
                <Box width={"100%"}>
                    <Checkbox
                        name='check'
                        checked={itemData.selected}
                        style={{ position: "absolute", zIndex: 2, boxShadow: "10px" }}
                        onChange={select}
                        onClick={(e) => { e.stopPropagation() }}
                    />
                </Box>

            )}
            <Box display={"flex"} justifyContent={"center"} alignItems={"center"} position={"absolute"} height={"100%"} width={"100%"} onClick={anyChecked ? () => { } : unselectedAction}>
                <ItemVisualComponent itemData={itemData} type={itemData.mediaData?.MediaType.FriendlyName} isDir={itemData.isDir} imported={itemData.imported} />
            </Box>
            <Box position={"absolute"} display={"flex"} justifyContent={"center"} alignItems={"center"} p={"10px"} bgcolor={"rgb(0, 0, 0, 0.50)"} width={"inherit"} height={"max-content"} bottom={"0px"} textAlign={"center"}>
                <TextBox itemData={itemData} editing={editing} hasInfo={hasInfo} setRenameVal={setRenameVal} dispatch={dispatch} />
            </Box>
        </Box>
    )
}