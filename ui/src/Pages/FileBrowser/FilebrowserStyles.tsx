import { AspectRatio, Box, Card, CardContent, Sheet, Typography, styled, useTheme } from "@mui/joy";
import { Dispatch } from "react";
import { HandleDrag } from "./FileBrowserLogic";
import { Folder } from "@mui/icons-material";

export const FlexColumnBox = styled(Box)({
    display: "flex",
    flexDirection: "column",
    alignItems: "center"
})

type DirViewWrapperProps = {
    path: string
    dragging: number
    hoverTarget: string
    dispatch: Dispatch<any>
    onDrop: (e: any) => void
    onMouseOver: (e: any) => void
    children: JSX.Element[]
}

export const DirViewWrapper = ({ path, dragging, hoverTarget, dispatch, onDrop, onMouseOver, children }: DirViewWrapperProps) => {
    const theme = useTheme()
    let dirName = path.slice(path.lastIndexOf('/', 1) + 1)
    if (dirName.endsWith('/')) {
        dirName = dirName.slice(0, -1)
    }

    if (dirName === "" || dirName === "/") {
        dirName = "Home"
    }

    return (
        <Box
            width={"100%"} height={'max-content'} display={'flex'} alignItems={'center'} justifyContent={'center'} paddingBottom={"100px"} paddingTop={"80px"}
            onDragOver={event => { HandleDrag(event, dispatch, dragging) }}
            onDrop={onDrop}
            onMouseOver={onMouseOver}
            onClick={() => { dispatch({ type: 'reject_edit' }); if (!dragging) { dispatch({ type: 'clear_selected' }) } else { dispatch({ type: 'set_dragging', dragging: false }) } }}
            // onMouseUp={() => { dispatch({ type: 'set_dragging', dragging: false }) }}
            zIndex={1}
        >
            {(dragging == 2) && (
                <Sheet
                    onDragLeave={event => { HandleDrag(event, dispatch, dragging) }}
                    sx={{
                        zIndex: 2,
                        bottom: "10px",
                        width: "calc(100% - 20px)",
                        height: "calc(100% - 100px)",
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
                                    {dirName}
                                </Typography>
                            </Typography>
                        </CardContent>
                    </Card>
                </Sheet>
            )}
            <Box
                display={"flex"}
                flexDirection={"column"}
                alignItems={"center"}
                width={"100%"}
                minHeight={"calc(100vh - 180px)"}
                height={"max-content"}
            >
                {children}
            </Box>
        </Box>
    )
}

export const DirItemsWrapper = styled(Box)({
    display: 'grid',
    gridGap: '16px',
    gridTemplateColumns: "repeat(auto-fill,minmax(200px,1fr))",

    paddingTop: "20px",

    width: '90%',
    maxWidth: '90%',
    height: 'max-content',

})

export const FileItemWrapper = ({ itemRef, hovering, isDir, selected, dragging, ...props }) => {
    const theme = useTheme()
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
        <Box ref={itemRef}>
            <Card
                {...props}
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
                <Box height={'100%'} width={'100%'} sx={{ backdropFilter: 'blur(2px)', backgroundColor: "#ffffff22", transform: 'translateY(-100%)', borderRadius: '10px' }} />

            )}
        </Box>
    )
}

export const ItemVisualComponentWrapper = ({ ...props }) => {
    return (
        <AspectRatio ratio={"1/1"} variant='solid' sx={{ width: "100%", ".MuiAspectRatio-content": { backgroundColor: 'transparent' } }}>
            <Box
                {...props}
                sx={{
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                    position: "static",
                    borderRadius: "10px",
                    overflow: 'hidden',
                    sx: { cursor: "pointer" },
                }}
            />
        </AspectRatio>
    )
}

export const TextBoxWrapper = ({ ...props }) => {
    const theme = useTheme()
    return (
        <Box
            {...props}
            sx={{
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