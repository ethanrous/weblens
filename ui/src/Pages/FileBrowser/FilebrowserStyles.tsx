import { Box, Card, MantineStyleProp, AspectRatio, Paper, Text, Tooltip, ActionIcon, Space } from '@mantine/core'
import { useCallback, useContext, useMemo, useRef, useState } from "react"
import { FilebrowserDragOver, HandleDrop } from "./FileBrowserLogic"
import { FileBrowserDispatch, fileData } from "../../types/Types"
import { useNavigate } from "react-router-dom"
import { IconFolder, IconFolderCancel, IconRefresh } from "@tabler/icons-react"
import { userContext } from '../../Context'
import '../../components/style.css'
import './style.css'


export const ColumnBox = ({ children, style, reff, className, onClick, onMouseOver, onMouseLeave, onContextMenu, onBlur, onDragOver, onDrop }: { children?, style?: MantineStyleProp, reff?, className?: string, onClick?, onMouseOver?, onMouseLeave?, onContextMenu?, onBlur?, onDragOver?, onDrop?}) => {
    return (
        <Box
            ref={reff}
            children={children}
            draggable={false}
            onClick={onClick}
            onMouseOver={onMouseOver}
            onMouseLeave={onMouseLeave}
            onContextMenu={onContextMenu}
            onBlur={onBlur}
            onDragOver={onDragOver}
            onDrop={onDrop}
            style={{
                display: "flex",
                height: "100%",
                width: "100%",
                flexDirection: "column",
                alignItems: 'center',
                // justifyContent: 'center',
                ...style,
            }}
            className={className}
        />
    )
}

export const RowBox = ({ children, style, onClick, onBlur }: { children, style?: MantineStyleProp, onClick?, onBlur?}) => {
    return (
        <Box
            children={children}
            onClick={onClick}
            onBlur={onBlur}
            style={{
                height: '100%',
                width: '100%',
                display: "flex",
                flexDirection: "row",
                alignItems: "center",
                ...style,
            }}
        />
    )
}


const Dropspot = ({ onDrop, dropspotTitle, dragging, dropAllowed, handleDrag }: { onDrop, dropspotTitle, dragging, dropAllowed, handleDrag: React.DragEventHandler<HTMLDivElement> }) => {
    return (
        <Box
            className='dropspot-wrapper'
            onDragOver={e => handleDrag(e)}
            style={{ pointerEvents: dragging === 2 ? 'all' : 'none', cursor: (!dropAllowed && dragging === 2) ? 'no-drop' : 'auto' }}
            onDragLeave={handleDrag}
        >
            {dragging === 2 && (
                <Box
                    onMouseLeave={handleDrag}
                    // onMouseMove={handleDrag}
                    className='dropbox'
                    onDrop={(e) => { console.log("here_again"); e.preventDefault(); e.stopPropagation(); dropAllowed && onDrop(e) }}
                    style={{ outlineColor: `${dropAllowed ? "#ffffff" : "#dd2222"}`, cursor: (!dropAllowed && dragging === 2) ? 'no-drop' : 'auto' }}
                >
                    {!dropAllowed && (
                        <ColumnBox style={{ position: 'relative', justifyContent: 'center', cursor: 'no-drop', width: 'max-content', pointerEvents: 'none' }}>
                            <IconFolderCancel size={100} color="#dd2222" />
                        </ColumnBox >
                    )}
                    {dropAllowed && (
                        <Card style={{ height: 'max-content', bottom: '20px', position: 'fixed' }}>
                            <RowBox>
                                <Text>
                                    Drop to upload to
                                </Text>
                                <IconFolder style={{ marginLeft: '7px' }} />
                                <Text fw={700} style={{ marginLeft: 3 }}>
                                    {dropspotTitle}
                                </Text>
                            </RowBox>
                        </Card>
                    )}
                </Box>
            )}
        </Box>
    )
}

type DirViewWrapperProps = {
    folderId: string
    folderName: string
    dragging: number
    dispatch: FileBrowserDispatch
    onDrop: (e: any) => void
    children: JSX.Element[]
}

export const DirViewWrapper = ({ folderId, folderName, dragging, dispatch, onDrop, children }: DirViewWrapperProps) => {
    const { userInfo } = useContext(userContext)
    const dropAllowed = useMemo(() => {
        return !(folderId === "shared" || folderId === userInfo.trashFolderId)
    }, [folderId, userInfo.trashFolderId])

    return (
        <Box
            style={{ height: "99%", width: "calc(100vw - (226px + 1vw))", position: 'relative' }}

            // If dropping is not allowed, and we drop, we want to clear the window when we detect the mouse moving again
            // We have to wait (a very short time, 10ms) to make sure the drop event fires and gets captured by the dropbox, otherwise
            // we set dragging to 0 too early, the dropbox gets removed, and chrome handles the drop event, opening the image in another tab
            onMouseMove={e => { if (dragging) { setTimeout(() => dispatch({ type: 'set_dragging', dragging: false }), 10) } }}
            onClick={() => dispatch({ type: "clear_selected" })}
        >
            <Dropspot onDrop={e => { dispatch({ type: 'set_dragging', dragging: false }); onDrop(e) }} dropspotTitle={folderName} dragging={dragging} dropAllowed={dropAllowed} handleDrag={event => { FilebrowserDragOver(event, dispatch, dragging) }} />
            <ColumnBox style={{ padding: 10 }} onDragOver={event => { FilebrowserDragOver(event, dispatch, dragging) }}>
                {children}
            </ColumnBox>
        </Box >
    )
}

export const WormholeWrapper = ({ wormholeId, wormholeName, validWormhole, uploadDispatch, children }: { wormholeId: string, wormholeName: string, validWormhole: boolean, uploadDispatch, children }) => {
    const { authHeader } = useContext(userContext)
    const [dragging, setDragging] = useState(0)
    const handleDrag = useCallback(e => { e.preventDefault(); if (e.type === "dragenter" || e.type === "dragover") { if (!dragging) { console.log("Yes drag"); setDragging(2) } } else if (dragging) { console.log("No drag"); setDragging(0) } }, [dragging])

    return (
        <Box className='wormhole-wrapper'>
            <Box
                style={{ position: 'relative', width: '98%', height: '98%' }}

                //                    See DirViewWrapper \/
                onMouseMove={e => { if (dragging) { setTimeout(() => setDragging(0), 10) } }}
            >

                <Dropspot
                    onDrop={(e => HandleDrop(e.dataTransfer.items, wormholeId, [], true, wormholeId, authHeader, uploadDispatch, () => { }))}
                    dropspotTitle={wormholeName}
                    dragging={dragging}
                    dropAllowed={validWormhole}
                    handleDrag={handleDrag} />
                <ColumnBox style={{ justifyContent: 'center' }} onDragOver={handleDrag}>
                    {children}
                </ColumnBox>
            </Box>
        </Box>

    )
}

export const ScanFolderButton = ({ folderId, holdingShift, doScan }) => {
    return (
        <Box >
            {folderId !== "shared" && folderId !== "trash" && (
                <Tooltip label={holdingShift ? "Deep scan folder" : "Scan folder"}>
                    <ActionIcon color='#00000000' size={35} onClick={doScan}>
                        <IconRefresh color={holdingShift ? '#4444ff' : 'white'} size={35} />
                    </ActionIcon>
                </Tooltip>
            )}
            {(folderId === "shared" || folderId === "trash") && (
                <Space w={35} />
            )}
        </Box>
    )
}

export const FileWrapper = ({ fileRef, fileData, width, dispatch, hovering, setHovering, isDir, selected, moveSelected, dragging, ...children }: { fileRef: any, fileData: fileData, width: number, dispatch: any, hovering: boolean, setHovering: any, isDir: boolean, selected: boolean, moveSelected: (entryId: string) => void, dragging: number, children: any }) => {
    const [mouseDown, setMouseDown] = useState(false)
    const navigate = useNavigate()

    const [outline, backgroundColor] = useMemo(() => {
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
        return [outline, backgroundColor]
    }, [selected, hovering, dragging, isDir])

    const MARGIN = 6

    return (
        <Box draggable={false} ref={fileRef} style={{ margin: MARGIN }}>
            <Card
                {...children}
                draggable={false}
                onClick={(e) => { e.stopPropagation(); dispatch({ type: 'set_selected', fileId: fileData.id }) }}
                onMouseOver={(e) => { e.stopPropagation(); setHovering(true); dispatch({ type: 'set_hovering', fileId: fileData.id }) }}
                onMouseUp={() => { if (dragging !== 0) { moveSelected(fileData.id) }; setMouseDown(false) }}
                onMouseDown={() => { setMouseDown(true) }}
                onDoubleClick={(e) => { e.stopPropagation(); if (fileData.isDir) { navigate(fileData.id) } }}
                onContextMenu={(e) => { e.preventDefault() }}
                onMouseLeave={() => {
                    setHovering(false)
                    if (!fileData.imported && !fileData.isDir) { return }
                    if (!selected && mouseDown) { dispatch({ type: "clear_selected" }) }
                    if (mouseDown) {
                        dispatch({ type: 'set_selected', fileId: fileData.id, selected: true })
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

                    height: (width - (MARGIN * 2)) * 1.10,
                    // height: '290px',
                    width: width - (MARGIN * 2),
                    // width: '250px',

                    // other
                    // position: 'relative',
                    cursor: (dragging !== 0 && !isDir) ? 'default' : 'pointer'
                }}
            />
            {(selected && dragging !== 0) && (
                <Box h={'100%'} w={'100%'} style={{ backgroundColor: "#ffffff22", transform: 'translateY(-100%)', borderRadius: '10px' }} />
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