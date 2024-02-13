import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react"
import { AspectRatio, Box, Divider, Text, Tooltip } from "@mantine/core"
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FilebrowserStyles"
import { MediaImage } from "./PhotoContainer"
import { MediaData } from "../types/Types"

type ItemMenu = ({ open, setOpen, itemInfo, menuPos }: { open: boolean, setOpen: (o: boolean) => void, itemInfo: ItemProps, menuPos: { x: number, y: number } }) => JSX.Element

export type GlobalContextType = {
    visitItem: (itemId: string) => void
    setDragging: (d: boolean) => void
    blockFocus: (b: boolean) => void
    rename: (itemId: string, newName: string) => void
    menu: ItemMenu

    setSelected?: (itemId: string, selected?: boolean) => void
    selectAll?: (itemId: string, selected?: boolean) => void
    moveSelected?: (itemId: string) => void

    iconDisplay?: ({ itemInfo }: { itemInfo: ItemProps }) => JSX.Element

    dragging?: number
    numCols?: number
    itemWidth?: number
    initialScrollIndex?: number
}

export type ItemProps = {
    itemId: string
    itemTitle: string
    secondaryInfo?: string
    selected: boolean
    mediaData: MediaData
    droppable: boolean
    isDir: boolean
    imported: boolean
    // useKeyDown?: () => void
    // moveSelected?: (entryId: string) => void,
    dragging?: number
    dispatch?: any

    extras?: any
}

type WrapperProps = {
    itemInfo: ItemProps
    fileRef

    width: number
    children

    visitItem
    setSelected: (itemId: string, selected?: boolean) => void
    moveSelected: (entryId: string) => void,
    ItemMenu: ItemMenu

    dragging: number// Allows for multiple dragging states
    setDragging: (d: boolean) => void
}

type TitleProps = {
    blockFocus: (b: boolean) => void
    rename: (itemId: string, newName: string) => void
    itemId: string
    itemTitle: string
    secondaryInfo?: string
    height: number
}

const MARGIN = 6
// useKeyDown(dispatch, editing, setEditingPlus, fileData.parentFolderId, fileData.id, renameVal, fileData.imported, authHeader)

const ItemWrapper = ({ itemInfo, fileRef, width, setSelected, dragging = 0, setDragging, visitItem, moveSelected, ItemMenu, children }: WrapperProps) => {
    const [mouseDown, setMouseDown] = useState(false)
    const [hovering, setHovering] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })

    const [outline, backgroundColor] = useMemo(() => {
        let outline
        let backgroundColor
        if (itemInfo.selected) {
            outline = `1px solid #220088`
            backgroundColor = "#331177"
        } else if (hovering && dragging && itemInfo.isDir) {
            outline = `2px solid #661199`
        } else if (hovering && !dragging) {
            backgroundColor = "#333333"
        } else {
            backgroundColor = "#222222"
        }
        return [outline, backgroundColor]
    }, [itemInfo.selected, hovering, dragging, itemInfo.isDir])

    return (
        <Box draggable={false} ref={fileRef} style={{ margin: MARGIN }}>
            <ItemMenu open={menuOpen} setOpen={setMenuOpen} itemInfo={itemInfo} menuPos={menuPos} />
            <Box
                draggable={false}
                children={children}
                onClick={(e) => { e.stopPropagation(); setSelected(itemInfo.itemId) }}
                onMouseOver={(e) => { e.stopPropagation(); setHovering(true) }}
                onMouseUp={() => { if (dragging !== 0) { moveSelected(itemInfo.itemId) }; setMouseDown(false) }}
                onMouseDown={() => { setMouseDown(true) }}
                onDoubleClick={(e) => { e.stopPropagation(); visitItem(itemInfo.itemId); }}
                onContextMenu={(e) => { e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true) }}
                onMouseLeave={() => {
                    setHovering(false)
                    // if (!fileData.imported && !fileData.isDir) { return }
                    // if (!selected && mouseDown) { dispatch({ type: "clear_selected" }) }
                    if (mouseDown) {
                        setDragging(true)
                        setSelected(itemInfo.itemId, true)
                        setMouseDown(false)
                    }
                }}
                variant='solid'
                style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    overflow: 'hidden',
                    borderRadius: '10px',
                    justifyContent: 'center',
                    outline: outline,
                    backgroundColor: backgroundColor,
                    padding: 1,
                    height: (width - (MARGIN * 2)) * 1.10,
                    width: width - (MARGIN * 2),
                    cursor: (dragging !== 0 && !itemInfo.isDir) ? 'default' : 'pointer'
                }}
            />
            {(itemInfo.selected && dragging !== 0) && (
                <Box style={{ height: (width - (MARGIN * 2)) * 1.10, width: width - (MARGIN * 2), position: 'absolute', backgroundColor: "#77777744", transform: 'translateY(-100%)', borderRadius: '10px' }} />
            )}
        </Box>
    )
}

export const FileVisualWrapper = ({ children }) => {
    return (
        <AspectRatio ratio={1} w={"94%"} display={'flex'} m={'6px'}>
            <Box children={children} style={{ overflow: 'hidden', borderRadius: '5px' }} />
        </AspectRatio>
    )
}

const useKeyDown = (itemId, newName, editing, setEditing, rename) => {

    const onKeyDown = useCallback((event) => {
        if (!editing) { return }
        if (event.key === "Enter") {
            rename(itemId, newName)
            setEditing(false)
        } else if (event.key === "Escape") {
            setEditing(false)

            // Rename with empty name is a "cancel" to the rename
            rename(itemId, "")
        }
    }, [itemId, newName, editing, setEditing, rename])

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [onKeyDown])
}

const TextBox = ({ itemId, itemTitle, secondaryInfo, height, blockFocus, rename }: TitleProps) => {
    const editRef: React.Ref<HTMLInputElement> = useRef()
    const [editing, setEditing] = useState(itemId === "")
    const [renameVal, setRenameVal] = useState(itemTitle)

    const setEditingPlus = useCallback((b: boolean) => { setEditing(b); setRenameVal(cur => { if (cur === '') { return itemTitle } else { return cur } }); blockFocus(b) }, [itemTitle, setEditing, blockFocus])
    useKeyDown(itemId, renameVal, editing, setEditingPlus, rename)

    useEffect(() => {
        if (editRef.current) {
            editRef.current.select()
        }
    }, [editing])

    useEffect(() => {
        if (itemId === "") {
            console.log("He")
            setEditingPlus(true)
        }
    }, [itemId, setEditingPlus])

    if (editing) {
        return (
            <ColumnBox style={{ height: height, width: '100%', justifyContent: 'center', alignItems: 'center' }} onBlur={() => { rename(itemId, ""); setEditingPlus(false) }}>
                <input
                    ref={editRef}
                    defaultValue={itemTitle}
                    onClick={(e) => { e.stopPropagation() }}
                    onDoubleClick={(e) => { e.stopPropagation() }}
                    onChange={(e) => { setRenameVal(e.target.value) }}
                    style={{ width: '90%', backgroundColor: '#00000000', border: 0, outline: 0 }}
                />
            </ColumnBox>
        )
    } else {
        return (
            <ColumnBox style={{ height: height, width: '100%', cursor: 'text', justifyContent: 'center', alignItems: 'center', paddingBottom: MARGIN / 2 }} onClick={(e) => { e.stopPropagation(); setEditingPlus(true) }}>
                <RowBox style={{ justifyContent: 'space-between', width: '90%', height: '90%' }}>
                    <Text size={`${height - (MARGIN * 2)}px`} truncate={'end'} style={{ color: "white", userSelect: 'none', lineHeight: 1.5 }}>{itemTitle}</Text>
                    <Divider orientation='vertical' my={1} mx={6} />
                    <ColumnBox style={{ width: 'max-content', justifyContent: 'center' }}>
                        <Text size={`${height - (MARGIN * 2 + 4)}px`} lineClamp={1} style={{ color: "white", overflow: 'visible', userSelect: 'none', width: 'max-content' }}> {secondaryInfo} </Text>
                    </ColumnBox>
                </RowBox>
                <Tooltip openDelay={300} label={itemTitle}>
                    <Box style={{ position: 'absolute', width: '90%', height: height }} />
                </Tooltip>
            </ColumnBox>
        )
    }
}

export const ItemDisplay = memo(({ itemInfo, context }: { itemInfo: ItemProps, context: GlobalContextType }) => {
    const wrapRef = useRef()

    const imgStyle = useMemo(() => {
        if (!itemInfo.mediaData) {
            return
        }
        let imgStyle
        if (itemInfo.mediaData.mediaHeight > itemInfo.mediaData.mediaWidth) {
            imgStyle = { width: '100%', height: 'auto' }
        } else {
            imgStyle = { height: '100%', width: 'auto' }
        }
        return imgStyle
    }, [itemInfo.mediaData])

    return (
        <ItemWrapper
            itemInfo={itemInfo}
            fileRef={wrapRef}
            setSelected={context.setSelected}
            visitItem={context.visitItem}
            width={context.itemWidth}
            moveSelected={context.moveSelected}
            dragging={context.dragging}
            setDragging={context.setDragging}
            ItemMenu={context.menu}
        >
            <FileVisualWrapper>
                {itemInfo.mediaData && (
                    <MediaImage media={itemInfo.mediaData} quality="thumbnail" imgStyle={imgStyle} />
                )}
                {!itemInfo.mediaData && context.iconDisplay && (
                    <context.iconDisplay itemInfo={itemInfo} />
                )}
            </FileVisualWrapper>

            <TextBox itemId={itemInfo.itemId} itemTitle={itemInfo.itemTitle} secondaryInfo={itemInfo.secondaryInfo} height={context.itemWidth * 0.10} blockFocus={context.blockFocus} rename={context.rename} />
        </ItemWrapper>
    )
}, (prev, next) => {
    if (prev.context !== next.context) {
        return false
    } else if (prev.context.itemWidth !== next.context.itemWidth) {
        return false
    } else if (prev.context.dragging !== next.context.dragging) {
        return false
    } else if (prev.itemInfo.selected !== next.itemInfo.selected) {
        return false
    }
    return true
})