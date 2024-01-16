import { useEffect, useState, memo, useRef, useCallback } from 'react'

import { humanFileSize } from '../../util'
import { CreateFolder, RenameFile } from '../../api/FileBrowserApi'
import { FileItemWrapper, FlexColumnBox, FlexRowBox, ItemVisualComponentWrapper } from './FilebrowserStyles'
import { fileData } from '../../types/Types'

import { MediaImage } from '../../components/PhotoContainer'
import { IconFile, IconFileZip, IconFolder } from '@tabler/icons-react'
import { Box, Center, Divider, Skeleton, Text, TextInput, Tooltip } from '@mantine/core'

function useKeyDown(dispatch, editing, setEditing, parentId, itemId, newName, imported, authHeader) {
    const keyDownHandler = useCallback(event => {
        if (!editing) { return }
        if (event.key === 'Enter') {
            event.preventDefault()
            event.stopPropagation()
            if (newName === "") {
                setEditing(false)
            } else {
                if (imported) {
                    RenameFile(itemId, newName, authHeader).then(() => setEditing(false))
                } else {
                    CreateFolder(parentId, newName, authHeader).then(folderId => {
                        dispatch({ type: 'set_selected', itemId: folderId })
                        setEditing(false)
                    })
                }
            }
        } else if (event.key === 'Escape') {
            setEditing(false)
        }
    }, [dispatch, editing, setEditing, itemId, newName, imported, authHeader, parentId])

    useEffect(() => {
        window.addEventListener('keydown', keyDownHandler)
        return () => { window.removeEventListener('keydown', keyDownHandler) }
    }, [keyDownHandler])
}

const ItemVisualComponent = ({ itemData, root }: { itemData: fileData, root }) => {
    const sqareSize = "75%"
    const type = itemData.mediaData?.mediaType.FriendlyName
    if (itemData.isDir) {
        return (<IconFolder style={{ width: sqareSize, height: sqareSize }} />)
    } else if (itemData.mediaData?.mediaType.IsDisplayable && itemData.imported) {
        return (<MediaImage mediaId={itemData.mediaData.fileHash} blurhash={itemData.mediaData.blurHash} metadataPreload={itemData.mediaData} quality={"thumbnail"} lazy root={root} />)
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

const Item = memo(({ itemData, selected, root, moveSelected, dragging, dispatch, authHeader }: {
    itemData: fileData, selected: boolean, root, moveSelected: () => void, dragging: number, dispatch: any, authHeader: any
}) => {
    const [hovering, setHovering] = useState(false)
    const [editing, setEditing] = useState(false)
    const [renameVal, setRenameVal] = useState(itemData.filename)
    const itemRef = useRef()

    const setEditingPlus = useCallback((b: boolean) => {setEditing(b); setRenameVal(cur => {if (cur === '') {return itemData.filename} else {return cur}}); dispatch({type: 'set_block_focus', block: b})}, [setEditing, dispatch])
    useKeyDown(dispatch, editing, setEditingPlus, itemData.parentFolderId, itemData.id, renameVal, itemData.imported, authHeader)

    useEffect(() => {
        if (itemData.id === "TEMPLATE_NEW_FOLDER") {
            setEditingPlus(true)
        }
    }, [itemData.id, setEditingPlus])

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
                <ItemVisualComponent itemData={itemData} root={root} />
            </ItemVisualComponentWrapper>

            <TextBox filename={itemData.filename} fileId={itemData.id} fileSize={itemData.size} editing={editing} setEditing={setEditingPlus} renameVal={renameVal} setRenameVal={setRenameVal} dispatch={dispatch} />
        </FileItemWrapper>
    )
}, (prev, next) => {
    // if (prev.itemData.visible !== next.itemData.visible) {
    //     return false
    // }

    // if (!next.itemData.visible) {
    //     return true
    // }

    if (prev.selected !== next.selected) {
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