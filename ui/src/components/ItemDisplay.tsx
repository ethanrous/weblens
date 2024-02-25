import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react"
import { AspectRatio, Box, Divider, Loader, Text, Tooltip } from "@mantine/core"
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FilebrowserStyles"
import { MediaImage } from "./PhotoContainer"
import { MediaData } from "../types/Types"
import "./filebrowserStyle.css"
import { IconUsersGroup } from "@tabler/icons-react"

type ItemMenu = ({ open, setOpen, itemInfo, menuPos }: { open: boolean, setOpen: (o: boolean) => void, itemInfo: ItemProps, menuPos: { x: number, y: number } }) => JSX.Element

export type GlobalContextType = {
    visitItem: (itemId: string) => void
    setDragging: (d: boolean) => void
    blockFocus: (b: boolean) => void
    rename: (itemId: string, newName: string) => void

    setMenuOpen: (o: boolean) => void,
    setMenuPos: ({ x, y }: { x: number, y: number }) => void
    setMenuTarget: (itemId: string) => void

    setSelected?: (itemId: string, selected?: boolean) => void
    selectAll?: (itemId: string, selected?: boolean) => void
    moveSelected?: (itemId: string) => void

    iconDisplay?: ({ itemInfo }: { itemInfo: ItemProps }) => JSX.Element
    setMoveDest?: (itemName) => void

    dragging?: number
    numCols?: number
    itemWidth?: number
    initialScrollIndex?: number
    doMediaFetch?: boolean
}

export type ItemProps = {
    itemId: string
    itemTitle: string
    secondaryInfo?: string
    selected: number
    mediaData: MediaData
    droppable: boolean
    isDir: boolean
    imported: boolean
    displayable: boolean
    dragging?: number
    dispatch?: any
    shares?: any[]

    extraIcons?: any[]

}

type WrapperProps = {
    itemInfo: ItemProps
    fileRef

    width: number
    editing: boolean
    children

    visitItem
    setSelected: (itemId: string, selected?: boolean) => void
    moveSelected: (entryId: string) => void,
    setMoveDest: (itemName: string) => void,

    dragging: number// Allows for multiple dragging states
    setDragging: (d: boolean) => void

    setMenuOpen: (o: boolean) => void,
    setMenuPos: ({ x, y }: { x: number, y: number }) => void
    setMenuTarget: (itemId: string) => void
}

type TitleProps = {
    itemId: string
    itemTitle: string
    secondaryInfo?: string
    editing: boolean
    setEditing: (e: boolean) => void
    height: number
    blockFocus: (b: boolean) => void
    rename: (itemId: string, newName: string) => void
}

const MARGIN = 6

const ItemWrapper = memo(({ itemInfo, fileRef, width, editing, setSelected, dragging = 0, setDragging, visitItem, moveSelected, setMenuOpen, setMenuPos, setMenuTarget, setMoveDest, children }: WrapperProps) => {
    const [mouseDown, setMouseDown] = useState(null)
    const [hovering, setHovering] = useState(false)

    const [outline, backgroundColor] = useMemo(() => {
        let outline
        let backgroundColor = "#222222"
        if (itemInfo.selected) {
            if (itemInfo.selected & 0x01) {
                backgroundColor = "#331177"
            }
            if (itemInfo.selected & 0x10) {
                outline = `1px solid #777777`
            }
        } else if (hovering && dragging && itemInfo.isDir) {
            outline = `2px solid #661199`
        } else if (hovering && !dragging) {
            backgroundColor = "#333333"
        }
        return [outline, backgroundColor]
    }, [itemInfo.selected, hovering, dragging, itemInfo.isDir])

    return (
        <Box
            ref={fileRef}
            style={{ margin: MARGIN }}
            onMouseOver={e => { e.stopPropagation(); setHovering(true); if (dragging && !itemInfo.selected && itemInfo.isDir) { setMoveDest(itemInfo.itemTitle) } }}
            onMouseDown={e => { setMouseDown({ x: e.clientX, y: e.clientY }) }}
            onMouseMove={e => { if (mouseDown && !dragging && (Math.abs(mouseDown.x - e.clientX) > 20 || Math.abs(mouseDown.y - e.clientY) > 20)) { setSelected(itemInfo.itemId, true); setDragging(true) } }}
            onClick={e => { e.stopPropagation(); setSelected(itemInfo.itemId) }}
            onMouseUp={e => {
                e.stopPropagation();
                if (dragging !== 0) {
                    if (!itemInfo.selected && itemInfo.isDir) {
                        moveSelected(itemInfo.itemId)
                    }
                    setMoveDest("")
                    setDragging(false)
                }
                setMouseDown(null)
            }}
            onDoubleClick={e => { e.stopPropagation(); visitItem(itemInfo.itemId) }}
            onContextMenu={e => {
                e.preventDefault();
                e.stopPropagation();

                setMenuTarget(itemInfo.itemId)
                setMenuPos({ x: e.clientX, y: e.clientY })
                setMenuOpen(true)
            }}
            onMouseLeave={e => {
                setHovering(false)
                if (dragging && itemInfo.isDir) { setMoveDest("") }
                if (mouseDown) { setMouseDown(null) }
            }}
        >
            <Box
                className="item-child"
                children={children}
                style={{
                    outline: outline,
                    backgroundColor: backgroundColor,
                    height: (width - (MARGIN * 2)) * 1.10,
                    width: width - (MARGIN * 2),
                    cursor: (dragging !== 0 && !itemInfo.isDir) ? 'default' : 'pointer',
                }}
            />
            {((itemInfo.selected || !itemInfo.isDir) && dragging !== 0) && (
                <Box
                    className="no-drop-cover"
                    style={{ height: (width - (MARGIN * 2)) * 1.10, width: width - (MARGIN * 2) }}
                    onMouseLeave={e => setHovering(false)}
                    onClick={e => e.stopPropagation()}
                />
            )}
        </Box>
    )
}, (prev, next) => {
    if (prev.itemInfo !== next.itemInfo) {
        return false
    } else if (prev.itemInfo.selected !== next.itemInfo.selected) {
        return false
    } else if (prev.editing !== next.editing) {
        return false
    } else if (prev.dragging !== next.dragging) {
        return false
    } else if (prev.width !== next.width) {
        return false
    }
    return true
})

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

const TextBox = ({ itemId, itemTitle, secondaryInfo, editing, setEditing, height, blockFocus, rename }: TitleProps) => {
    const editRef: React.Ref<HTMLInputElement> = useRef()
    const [renameVal, setRenameVal] = useState(itemTitle)

    const setEditingPlus = useCallback((b: boolean) => { setEditing(b); setRenameVal(cur => { if (cur === '') { return itemTitle } else { return cur } }); blockFocus(b) }, [itemTitle, setEditing, blockFocus])
    useKeyDown(itemId, renameVal, editing, setEditingPlus, rename)

    useEffect(() => {
        if (editing && editRef.current) {
            editRef.current.select()
        }
    }, [editing, editRef])

    useEffect(() => {
        if (itemId === "NEW_DIR") {
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
    const [editing, setEditing] = useState(false)

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
            setMoveDest={context.setMoveDest}

            setMenuOpen={context.setMenuOpen}
            setMenuPos={context.setMenuPos}
            setMenuTarget={context.setMenuTarget}

            editing={editing}
        >
            <FileVisualWrapper>
                {itemInfo.mediaData && (
                    <MediaImage media={itemInfo.mediaData} quality="thumbnail" doFetch={context.doMediaFetch} />
                )}
                {!itemInfo.mediaData && context.iconDisplay && (
                    <context.iconDisplay itemInfo={itemInfo} />
                )}
                <RowBox style={{ position: 'absolute', alignItems: 'flex-start', padding: 5 }}>
                    {itemInfo.extraIcons?.map((Icon, i) => (
                        <Icon key={i} style={{ filter: 'drop-shadow(1px 2px 1.5px black)' }} />

                    ))}
                </RowBox>
            </FileVisualWrapper>

            <TextBox itemId={itemInfo.itemId} itemTitle={itemInfo.itemTitle} secondaryInfo={itemInfo.secondaryInfo} editing={editing} setEditing={setEditing} height={context.itemWidth * 0.10} blockFocus={context.blockFocus} rename={context.rename} />

            {itemInfo.itemId === "NEW_DIR" && !editing && (
                <Loader color="white" size={20} style={{ position: 'absolute', top: 20, right: 20 }} />
            )}
        </ItemWrapper>
    )
}, (prev, next) => {
    if (prev.itemInfo.itemId !== next.itemInfo.itemId) {
        return false
    } else if (prev.context !== next.context) {
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