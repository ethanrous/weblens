import { useEffect, useState, memo, useRef, useCallback } from 'react'

import { humanFileSize } from '../../util'
import { CreateFolder, RenameFile } from '../../api/FileBrowserApi'
import { FileVisualWrapper, FileWrapper, FlexColumnBox, FlexRowBox } from './FilebrowserStyles'
import { fileData } from '../../types/Types'

import { MediaImage } from '../../components/PhotoContainer'
import { IconFile, IconFileZip, IconFolder } from '@tabler/icons-react'
import { Box, Center, Divider, Skeleton, Text, TextInput, Tooltip } from '@mantine/core'

function useKeyDown(dispatch, editing, setEditing, parentId, fileId, newName, imported, authHeader) {
    const keyDownHandler = useCallback(event => {
        if (!editing) { return }
        if (event.key === 'Enter') {
            event.preventDefault()
            event.stopPropagation()
            if (newName === "") {
                setEditing(false)
            } else {
                if (imported) {
                    RenameFile(fileId, newName, authHeader).then(() => setEditing(false))
                } else {
                    dispatch({type: 'set_loading', loading: true})
                    dispatch({ type: 'delete_from_map', fileId: "TEMPLATE_NEW_FOLDER" })
                    CreateFolder(parentId, newName, authHeader).then(folderId => {
                        dispatch({ type: 'set_selected', fileId: folderId })
                        setEditing(false)
                        dispatch({type: 'set_loading', loading: false})
                    })
                }
            }
        } else if (event.key === 'Escape') {
            setEditing(false)
        }
    }, [dispatch, editing, setEditing, fileId, newName, imported, authHeader, parentId])

    useEffect(() => {
        window.addEventListener('keydown', keyDownHandler)
        return () => { window.removeEventListener('keydown', keyDownHandler) }
    }, [keyDownHandler])
}

const FileVisualComponent = ({ fileData, root }: { fileData: fileData, root }) => {
    const sqareSize = "75%"
    const type = fileData.mediaData?.mediaType?.FriendlyName
    if (fileData.isDir) {
        return (<IconFolder style={{ width: sqareSize, height: sqareSize }} />)
    } else if (fileData.displayable && fileData.imported) {
        return (<MediaImage mediaId={fileData?.mediaData?.fileHash} blurhash={fileData?.mediaData?.blurHash} metadataPreload={fileData.mediaData} quality={"thumbnail"} lazy root={root} />)
    } else if (type === "File") {
        return (<IconFile style={{ width: sqareSize, height: sqareSize }} />)
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

const TextBox = ({ filename, fileId, fileSize, editing, setEditing, renameVal, setRenameVal, dispatch }: { filename: string, fileId: string, fileSize: number, editing: boolean, setEditing: (boolean) => void, renameVal: string, setRenameVal: any, dispatch: any }) => {
    const editRef: React.Ref<HTMLInputElement> = useRef()
    useEffect(() => {
        if (editRef.current) {
            editRef.current.select()
        }
    }, [editing])

    if (editing) {
        return (
            <FlexColumnBox style={{ justifyContent: 'center', height: '40px' }} onBlur={() => setEditing(false)}>
                <TextInput
                    autoFocus
                    ref={editRef}
                    w={'90%'}
                    variant='unstyled'
                    error={renameVal === ""}
                    defaultValue={filename}
                    onClick={(e) => { e.stopPropagation() }}
                    onDoubleClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}
                />
            </FlexColumnBox>
        )
    } else {
        const [sizeValue, units] = humanFileSize(fileSize, true)
        return (
            <FlexColumnBox style={{ width: '100%', cursor: 'text', padding: '5px' }} onClick={(e) => { e.stopPropagation(); setEditing(true) }}>
                <FlexRowBox style={{ justifyContent: 'space-evenly', width: '90%', height: '30px' }}>
                    <Text size={'15px'} truncate={'end'} style={{ color: "white", userSelect: 'none', lineHeight: 1.5 }}>{filename}</Text>
                    <Divider orientation='vertical' m={6} />
                    <FlexColumnBox style={{ width: 'max-content', justifyContent: 'center' }}>
                        <Text size={'10px'} style={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {sizeValue} </Text>
                        <Text size={'10px'} style={{ color: "white", overflow: 'visible', userSelect: 'none' }}> {units} </Text>
                    </FlexColumnBox>
                </FlexRowBox>
                <Tooltip openDelay={300} label={filename}>
                    <Box style={{ position: 'absolute', width: '100%', height: '100%' }} />
                </Tooltip>
            </FlexColumnBox>
        )
    }
}

const File = memo(({ fileData, selected, root, moveSelected, dragging, dispatch, authHeader }: {
    fileData: fileData, selected: boolean, root, moveSelected: () => void, dragging: number, dispatch: any, authHeader: any
}) => {
    const [hovering, setHovering] = useState(false)
    const [editing, setEditing] = useState(false)
    const [renameVal, setRenameVal] = useState(fileData.filename)
    const fileRef = useRef()

    const setEditingPlus = useCallback((b: boolean) => {setEditing(b); setRenameVal(cur => {if (cur === '') {return fileData.filename} else {return cur}}); dispatch({type: 'set_block_focus', block: b})}, [setEditing, dispatch])
    useKeyDown(dispatch, editing, setEditingPlus, fileData.parentFolderId, fileData.id, renameVal, fileData.imported, authHeader)

    useEffect(() => {
        if (fileData.id === "TEMPLATE_NEW_FOLDER") {
            setEditingPlus(true)
        }
    }, [fileData.id, setEditingPlus])

    return (
        <FileWrapper
            fileRef={fileRef}
            fileData={fileData}
            dispatch={dispatch}
            hovering={hovering}
            setHovering={setHovering}
            isDir={fileData.isDir}
            selected={selected}
            moveSelected={moveSelected}
            dragging={dragging}
        >
            <FileVisualWrapper>
                <FileVisualComponent fileData={fileData} root={root} />
            </FileVisualWrapper>

            <TextBox filename={fileData.filename} fileId={fileData.id} fileSize={fileData.size} editing={editing} setEditing={setEditingPlus} renameVal={renameVal} setRenameVal={setRenameVal} dispatch={dispatch} />
        </FileWrapper>
    )
}, (prev, next) => {
    // if (prev.fileData.visible !== next.fileData.visible) {
    //     return false
    // }

    // if (!next.fileData.visible) {
    //     return true
    // }

    if (prev.selected !== next.selected) {
        return false
    } else if (prev.dragging !== next.dragging) {
        return false
    } else if (prev.fileData.imported !== next.fileData.imported) {
        return false
    } else if (prev.fileData.size !== next.fileData.size) {
        return false
    }
    return true
})

export default File