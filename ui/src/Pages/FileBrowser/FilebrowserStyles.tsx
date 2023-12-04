import { Card, CardContent, Sheet, Typography, styled, useTheme } from "@mui/joy"
import { Box, MantineStyleProp, rem } from '@mantine/core'
import { Dispatch, useState } from "react"
import { HandleDrag } from "./FileBrowserLogic"
import { Folder } from "@mui/icons-material"
import { AspectRatio, Tooltip } from "@mantine/core"
import { itemData } from "../../types/Types"
import { useNavigate } from "react-router-dom"

export const FlexColumnBox = ({ children, style }: { children, style?: MantineStyleProp }) => {
    return (
        <Box
            children={children}
            style={{
                ...style,
                display: "flex",
                flexDirection: "column",
                alignItems: "center"
            }}
        />
    )
}

export const FlexRowBox = ({ children, ...style }) => {
    return (
        <Box
            children={children}
            w={"100%"}
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
    const theme = useTheme()

    return (
        <Box
            w={"100%"} display={'flex'} pt={"80px"}
            style={{ zIndex: 1, height: "calc(100vh - 20px)" }}
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
                        outline: `1px solid ${theme.colorSchemes.dark.palette.primary.outlinedColor}`
                    }}
                >
                    <Card variant="plain" orientation="horizontal" sx={{ height: 'max-content', bottom: '20px', position: 'fixed' }}>
                        <CardContent>
                            <Typography level="title-md" display={'flex'}>
                                {"Drop to upload to"}
                                <Folder sx={{ marginLeft: '7px' }} />
                                <Typography fontWeight={'lg'} marginLeft={'3px'}>
                                    {folderName}
                                </Typography>
                            </Typography>
                        </CardContent>
                    </Card>
                </Sheet>
            )}
            <Box
                display={"flex"}
                pl={20}
                w={"100%"}
                style={{ flexDirection: 'column' }}
            >
                {children}
            </Box>

        </Box>
    )
}

export const DirItemsWrapper = ({ children }) => {
    return (
        <Box
            children={children}
            style={{
                display: 'grid',
                gridGap: '16px',
                gridTemplateColumns: "repeat(auto-fill,minmax(200px,1fr))",
                touchAction: 'none',
                overflowY: 'scroll',
                borderRadius: '10px',
                paddingTop: "1px",
                paddingBottom: "1px",
                paddingLeft: "1px",
                paddingRight: "5vw",
                width: "105%"
            }}
        />
    )
}

export const FileItemWrapper = ({ itemRef, itemData, dispatch, hovering, setHovering, isDir, selected, moveSelected, dragging, ...children }: { itemRef: any, itemData: itemData, dispatch: any, hovering: boolean, setHovering: any, isDir: boolean, selected: boolean, moveSelected: () => void, dragging: number, children: any }) => {
    const [mouseDown, setMouseDown] = useState(false)
    const theme = useTheme()
    const navigate = useNavigate()

    let outline
    let backgroundColor
    if (selected) {
        outline = `1px solid ${theme.colorSchemes.dark.palette.primary.outlinedColor}`
        backgroundColor = theme.colorSchemes.dark.palette.primary.solidActiveBg
    } else if (hovering && dragging && isDir) {
        outline = `2px solid ${theme.colorSchemes.dark.palette.primary.outlinedColor}`
    } else if (hovering && !dragging) {
        backgroundColor = theme.colorSchemes.dark.palette.primary.solidBg
    } else {
        backgroundColor = theme.colorSchemes.dark.palette.primary.solidDisabledBg
    }
    return (
        <Tooltip openDelay={500} label={itemData.filename}>
        <Box ref={itemRef}>
            <Card
                    {...children}
                    onClick={(e) => { e.stopPropagation(); dispatch({ type: 'reject_edit' }); if (!itemData.imported && !itemData.isDir) { return } dispatch({ type: 'set_selected', itemId: itemData.id }) }}
                    onMouseOver={(e) => { e.stopPropagation(); setHovering(true); dispatch({ type: 'set_hovering', itemId: itemData.id }) }}
                    onMouseUp={() => { if (dragging !== 0) { moveSelected() }; setMouseDown(false) }}
                    onMouseDown={() => { setMouseDown(true) }}
                    onDoubleClick={(e) => { e.stopPropagation(); if (itemData.isDir) { navigate(itemData.id) } else if (itemData.mediaData.MediaType.IsDisplayable) { dispatch({ type: 'set_presentation', presentingId: itemData.id }) } }}
                    onMouseLeave={() => {
                        setHovering(false);
                        if (!itemData.imported && !itemData.isDir) { return }
                        if (!selected && mouseDown) { dispatch({ type: "clear_selected" }) }
                        if (mouseDown) {
                            dispatch({ type: 'set_selected', itemId: itemData.id, selected: true })
                            dispatch({ type: 'set_dragging', dragging: true })
                            setMouseDown(false)
                        }
                    }}
                variant='solid'
                sx={{
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
                    cursor: 'pointer'
                }}
            />
                {(selected && dragging != 0) && (
                    <Box h={'100%'} w={'100%'} style={{ backdropFilter: 'blur(2px)', backgroundColor: "#ffffff22", transform: 'translateY(-100%)', borderRadius: '10px' }} />
            )}
        </Box>
        </Tooltip>
    )
}

export const ItemVisualComponentWrapper = ({ children }) => {
    return (
        <AspectRatio ratio={1} variant='solid' w={"100%"} display={'flex'}>
            <Box children={children} style={{ overflow: 'hidden', borderRadius: '5px' }} />
        </AspectRatio>
    )
}

export const TextBoxWrapper = ({ ...props }) => {
    const theme = useTheme()
    return (
        <Box
            {...props}
            style={{
                position: "relative",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                padding: "10px",
                backgroundColor: theme.colorSchemes.dark.palette.neutral.softBg,
                width: "100%",
                height: "max-content",
                bottom: "0px",
                textAlign: "center",
            }}
        />
    )
}