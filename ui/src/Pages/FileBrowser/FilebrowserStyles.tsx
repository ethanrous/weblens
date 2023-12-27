import { CardContent, Sheet, Typography } from "@mui/joy"
import { Box, Card, MantineStyleProp, AspectRatio } from '@mantine/core'
import { Dispatch, useState } from "react"
import { HandleDrag } from "./FileBrowserLogic"
import { itemData } from "../../types/Types"
import { useNavigate } from "react-router-dom"
import { IconFolder } from "@tabler/icons-react"

export const FlexColumnBox = ({ children, style, alignOverride, reff, onClick, onMouseOver, onMouseLeave, onContextMenu }: { children, style?: MantineStyleProp, alignOverride?: string, reff?, onClick?, onMouseOver?, onMouseLeave?, onContextMenu?}) => {
    let align = alignOverride || "center"
    return (
        <Box
            ref={reff}
            children={children}
            draggable={false}
            onClick={onClick}
            onMouseOver={onMouseOver}
            onMouseLeave={onMouseLeave}
            onContextMenu={onContextMenu}
            style={{
                display: "flex",
                flexDirection: "column",
                alignItems: align,
                ...style,
            }}
        />
    )
}

export const FlexRowBox = ({ children, style }: { children, style?: MantineStyleProp }) => {
    return (
        <Box
            children={children}
            style={{
                ...style,
                display: "flex",
                flexDirection: "row",
            }}
        />
    )
}

type DirViewWrapperProps = {
    folderName: string
    dragging: number
    hoverTarget: string
    dispatch: Dispatch<any>
    onDrop: (e: any) => void
    onMouseOver: (e: any) => void
    children: JSX.Element[]
}

export const DirViewWrapper = ({ folderName, dragging, hoverTarget, dispatch, onDrop, onMouseOver, children }: DirViewWrapperProps) => {

    return (
        <Box
            display={'flex'} pt={"80px"}
            style={{ zIndex: 1, height: "calc(100vh - 20px)", width: '100%' }}
            onDragOver={event => { HandleDrag(event, dispatch, dragging) }}
            onDrop={onDrop}
            onMouseOver={onMouseOver}
            onClick={() => { dispatch({ type: 'reject_edit' }); if (!dragging) { dispatch({ type: 'clear_selected' }) } else { dispatch({ type: 'set_dragging', dragging: false }) } }}
        >
            {(dragging == 2) && (
                <Sheet
                    onDragLeave={event => { HandleDrag(event, dispatch, dragging) }}
                    sx={{
                        zIndex: 2,
                        bottom: "10px",
                        left: "10px",
                        width: "calc(100% - 20px)",
                        height: "calc(100% - 90px)",
                        position: 'fixed',
                        display: 'flex',
                        flexDirection: 'row',
                        justifyContent: 'center',
                        backgroundColor: "#00000077",
                        backdropFilter: "blur(3px)",
                        outline: `1px solid white`
                    }}
                >
                    <Card style={{ height: 'max-content', bottom: '20px', position: 'fixed' }}>
                        <CardContent>
                            <Typography level="title-md" display={'flex'}>
                                {"Drop to upload to"}
                                <IconFolder style={{ marginLeft: '7px' }} />
                                <Typography fontWeight={'lg'} marginLeft={'3px'}>
                                    {folderName}
                                </Typography>
                            </Typography>
                        </CardContent>
                    </Card>
                </Sheet>
            )}
            <FlexColumnBox style={{ paddingLeft: 20, width: '100%' }}>
                {children}
            </FlexColumnBox>
        </Box>
    )
}

export const FileItemWrapper = ({ itemRef, itemData, dispatch, hovering, setHovering, isDir, selected, moveSelected, dragging, ...children }: { itemRef: any, itemData: itemData, dispatch: any, hovering: boolean, setHovering: any, isDir: boolean, selected: boolean, moveSelected: () => void, dragging: number, children: any }) => {
    const [mouseDown, setMouseDown] = useState(false)
    const navigate = useNavigate()

    let outline
    let backgroundColor
    if (selected) {
        outline = `1px solid #220088`
        backgroundColor = "#331177"
    } else if (hovering && dragging && isDir) {
        outline = `2px solid #661199`
    } else if (hovering && !dragging) {
        backgroundColor = "#333333"
    } else {
        backgroundColor = "#222222"
    }

    return (
        <Box draggable={false} ref={itemRef}>
            <Card
                {...children}
                draggable={false}
                onClick={(e) => { e.stopPropagation(); dispatch({ type: 'reject_edit' }); dispatch({ type: 'set_selected', itemId: itemData.id }) }}
                onMouseOver={(e) => { e.stopPropagation(); setHovering(true); dispatch({ type: 'set_hovering', itemId: itemData.id }) }}
                onMouseUp={() => { if (dragging !== 0) { moveSelected() }; setMouseDown(false) }}
                onMouseDown={() => { setMouseDown(true) }}
                onDoubleClick={(e) => { e.stopPropagation(); if (itemData.isDir) { navigate(itemData.id) } else if (itemData.mediaData.mediaType.IsDisplayable) { dispatch({ type: 'set_presentation', itemId: itemData.id }) } }}
                onContextMenu={(e) => { e.preventDefault() }}
                onMouseLeave={() => {
                    setHovering(false)
                    if (!itemData.imported && !itemData.isDir) { return }
                    if (!selected && mouseDown) { dispatch({ type: "clear_selected" }) }
                    if (mouseDown) {
                        dispatch({ type: 'set_selected', itemId: itemData.id, selected: true })
                        dispatch({ type: 'set_dragging', dragging: true })
                        setMouseDown(false)
                    }
                }}
                variant='solid'
                style={{
                    // internal
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    overflow: 'hidden',
                    borderRadius: '10px',
                    justifyContent: 'center',
                    outline: outline,
                    backgroundColor: backgroundColor,
                    padding: 1,

                    // sizing
                    maxWidth: '100%',
                    minWidth: '50%',
                    flexGrow: 1,
                    flexShrink: 1,
                    flexBasis: 200,

                    // other
                    position: 'relative',
                    cursor: (dragging !== 0 && !isDir) ? 'default' : 'pointer'
                }}
            />
            {(selected && dragging !== 0) && (
                <Box h={'100%'} w={'100%'} style={{ backgroundColor: "#ffffff22", transform: 'translateY(-100%)', borderRadius: '10px' }} />
            )}
        </Box>
    )
}

export const ItemVisualComponentWrapper = ({ children }) => {
    return (
        <AspectRatio ratio={1} w={"94%"} display={'flex'} m={'6px'}>
            <Box children={children} style={{ overflow: 'hidden', borderRadius: '5px' }} />
        </AspectRatio>
    )
}