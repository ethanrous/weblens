import { Box, Card, MantineStyleProp, Text, Tooltip, ActionIcon, Space, Menu, Divider, FileButton } from '@mantine/core'
import { useCallback, useContext, useMemo, useState } from "react"
import { FilebrowserDragOver, HandleDrop, HandleUploadButton } from "./FileBrowserLogic"
import { FileBrowserDispatch, fileData } from "../../types/Types"

import { IconFile, IconFileZip, IconFolder, IconFolderCancel, IconFolderPlus, IconRefresh, IconSpiral, IconUpload } from "@tabler/icons-react"
import { userContext } from '../../Context'
import '../../components/style.css'
import '../../components/filebrowserStyle.css'
import { humanFileSize } from '../../util'
import Crumbs from '../../components/Crumbs'


export const ColumnBox = ({ children, style, reff, className, onClick, onMouseOver, onMouseLeave, onContextMenu, onBlur, onDragOver }: { children?, style?: MantineStyleProp, reff?, className?: string, onClick?, onMouseOver?, onMouseLeave?, onContextMenu?, onBlur?, onDragOver?}) => {
    return (
        <Box
            ref={reff}
            children={children}
            onClick={onClick}
            onMouseOver={onMouseOver}
            onMouseLeave={onMouseLeave}
            onContextMenu={onContextMenu}
            onBlur={onBlur}
            onDragOver={onDragOver}
            // onDrop={onDrop}
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

export const TransferCard = ({ action, destination, boundRef }: { action: string, destination: string, boundRef?}) => {
    let width
    let left
    if (boundRef && boundRef.current) {
        width = boundRef.current.clientWidth
        left = boundRef.current.getBoundingClientRect()['left']
    }
    if (!destination) {
        return
    }

    return (
        <Box
            className='transfer-info-box'
            style={{ pointerEvents: 'none', width: width ? width : '100%', left: left ? left : 0 }}
        >
            <Card style={{ height: 'max-content' }}>
                <RowBox>
                    <Text>
                        {action} to
                    </Text>
                    <IconFolder style={{ marginLeft: '7px' }} />
                    <Text fw={700} style={{ marginLeft: 3 }}>
                        {destination}
                    </Text>
                </RowBox>
            </Card>
        </Box>
    )
}

const Dropspot = ({ onDrop, dropspotTitle, dragging, dropAllowed, handleDrag }: { onDrop, dropspotTitle, dragging, dropAllowed, handleDrag: React.DragEventHandler<HTMLDivElement> }) => {
    return (
        <Box
            className='dropspot-wrapper'
            onDragOver={e => { if (dragging === 0) { handleDrag(e) } }}
            style={{ pointerEvents: dragging === 2 ? 'all' : 'none', cursor: (!dropAllowed && dragging === 2) ? 'no-drop' : 'auto' }}
            onDragLeave={handleDrag}
        >
            {dragging === 2 && (
                <Box
                    className='dropbox'
                    onMouseLeave={handleDrag}
                    onDrop={e => { e.preventDefault(); e.stopPropagation(); if (dropAllowed) { onDrop(e) } }}

                    // required for onDrop to work https://stackoverflow.com/questions/50230048/react-ondrop-is-not-firing
                    onDragOver={e => e.preventDefault()}

                    style={{ outlineColor: `${dropAllowed ? "#ffffff" : "#dd2222"}`, cursor: (!dropAllowed && dragging === 2) ? 'no-drop' : 'auto' }}
                >
                    {!dropAllowed && (
                        <ColumnBox style={{ position: 'relative', justifyContent: 'center', cursor: 'no-drop', width: 'max-content', pointerEvents: 'none' }}>
                            <IconFolderCancel size={100} color="#dd2222" />
                        </ColumnBox >
                    )}
                    {dropAllowed && (
                        <TransferCard action='Upload' destination={dropspotTitle} />
                    )}
                </Box>
            )}
        </Box>
    )
}

const FilebrowserMenu = ({ folderName, menuPos, menuOpen, setMenuOpen, newFolder }) => {

    return (
        <Menu opened={menuOpen} onClose={() => setMenuOpen(false)}>
            <Menu.Target>
                <Box style={{ position: 'fixed', top: menuPos.y, left: menuPos.x }} />
            </Menu.Target>

            <Menu.Dropdown>
                <Menu.Label>{folderName}</Menu.Label>
                <Menu.Item leftSection={<IconFolderPlus />} onClick={() => newFolder()}>
                    New Folder
                </Menu.Item>

            </Menu.Dropdown>
        </Menu>
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
    const [menuOpen, setMenuOpen] = useState(false)
    const [menuPos, setMenuPos] = useState({ x: 0, y: 0 })
    const dropAllowed = useMemo(() => {
        return !(folderId === "shared" || folderId === userInfo.trashFolderId)
    }, [folderId, userInfo.trashFolderId])

    return (
        <Box
            style={{ height: "99%", width: "calc(100vw - (226px + 1vw))", position: 'relative' }}

            // If dropping is not allowed, and we drop, we want to clear the window when we detect the mouse moving again
            // We have to wait (a very short time, 10ms) to make sure the drop event fires and gets captured by the dropbox, otherwise
            // we set dragging to 0 too early, the dropbox gets removed, and chrome handles the drop event, opening the image in another tab
            // onMouseMove={e => { if (dragging) { setTimeout(() => dispatch({ type: 'set_dragging', dragging: false }), 10) } }}
            onMouseUp={e => { if (dragging) { setTimeout(() => dispatch({ type: 'set_dragging', dragging: false }), 10) } }}
            onClick={e => {
                if (dragging) {
                    return
                }
                dispatch({ type: "clear_selected" })
            }}
            onContextMenu={e => { e.preventDefault(); setMenuPos({ x: e.clientX, y: e.clientY }); setMenuOpen(true); }}
        >
            <FilebrowserMenu folderName={folderName === userInfo.username ? "Home" : folderName} menuPos={menuPos} menuOpen={menuOpen} setMenuOpen={setMenuOpen} newFolder={() => dispatch({ type: 'new_dir' })} />
            <Dropspot onDrop={e => { onDrop(e); dispatch({ type: 'set_dragging', dragging: false }) }} dropspotTitle={folderName} dragging={dragging} dropAllowed={dropAllowed} handleDrag={event => FilebrowserDragOver(event, dispatch, dragging)} />
            <ColumnBox style={{ padding: 10 }} onDragOver={event => { if (!dragging) { FilebrowserDragOver(event, dispatch, dragging) } }}>
                {children}
            </ColumnBox>
        </Box >
    )
}

export const WormholeWrapper = ({ wormholeId, wormholeName, validWormhole, uploadDispatch, children }: { wormholeId: string, wormholeName: string, validWormhole: boolean, uploadDispatch, children }) => {
    const { authHeader } = useContext(userContext)
    const [dragging, setDragging] = useState(0)
    const handleDrag = useCallback(e => { e.preventDefault(); if (e.type === "dragenter" || e.type === "dragover") { if (!dragging) { setDragging(2) } } else if (dragging) { setDragging(0) } }, [dragging])

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

export const FolderIcon = ({ shares, size = '75%' }: { shares, size?}) => {
    const [copied, setCopied] = useState(false)
    return (
        <RowBox style={{ justifyContent: 'center' }}>
            <IconFolder size={size} />
            {shares.length !== 0 && (
                <Tooltip label={copied ? 'Copied' : 'Copy Wormhole'}>
                    <IconSpiral
                        color={copied ? '#4444ff' : 'white'}
                        style={{ position: 'absolute', right: 0, top: 0 }}
                        onClick={e => { e.stopPropagation(); navigator.clipboard.writeText(`${window.location.origin}/wormhole/${shares[0].ShareId}`); setCopied(true); setTimeout(() => setCopied(false), 1000) }}
                        onDoubleClick={e => e.stopPropagation()}
                    />
                </Tooltip>
            )}
        </RowBox>
    )
}

const FileIcon = ({ file }: { file: fileData }) => {
    if (file.isDir) {
        return (<FolderIcon size={'60%'} shares={file.shares} />)
    }
    const ext = file.filename.slice(file.filename.indexOf('.') + 1, file.filename.length)

    switch (ext) {
        case "zip": return (<IconFileZip size={'60%'} />)
        default: return (
            <Box style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', width: '100%' }}>
                <IconFile size={'75%'} />
                <Text size='70px' fw={700} style={{ position: 'absolute', userSelect: 'none' }}>{ext}</Text>
            </Box>
        )
    }
}

export const PresentationFile = ({ file }: { file: fileData }) => {
    if (!file) {
        return null
    }
    let [size, units] = humanFileSize(file.size)
    if (file.displayable && file.mediaData) {
        return (() =>
            <ColumnBox style={{ justifyContent: 'center', width: '40%', height: 'max-content' }} onClick={e => e.stopPropagation()}>
                <Text fw={600} style={{ fontSize: '45px' }}>{file.filename}</Text>
                <Text style={{ fontSize: '25px' }}>{size}{units}</Text>
                <Text style={{ fontSize: '20px' }}>{new Date(Date.parse(file.mediaData.createDate)).toLocaleDateString('en-us', { year: "numeric", month: "short", day: "numeric" })}</Text>
            </ColumnBox>
        )
    } else {
        return (() =>
            <RowBox style={{ justifyContent: 'center', height: 'max-content' }} onClick={e => e.stopPropagation()}>
                <Box style={{ width: '60%', display: 'flex', justifyContent: 'center' }}>
                    <FileIcon file={file} />
                </Box>
                <Space w={30} />
                <ColumnBox style={{ width: '40%', justifyContent: 'center' }}>
                    <Text fw={600} style={{ fontSize: '45px' }}>{file.filename}</Text>
                    {file.isDir && (
                        <RowBox style={{ height: 'max-content', justifyContent: 'center', width: '50vw' }}>
                            <Text style={{ fontSize: '25px' }}>{file.children.length} Item{file.children.length !== 1 ? 's' : ''}</Text>
                            <Divider orientation="vertical" size={2} mx={10} />
                            <Text style={{ fontSize: '25px' }}>{size}{units}</Text>
                        </RowBox>
                    )}
                    {!file.isDir && (
                        <Text style={{ fontSize: '25px' }}>{size}{units}</Text>
                    )}
                </ColumnBox>
            </RowBox>
        )
    }
}

export const GetStartedCard = ({ filebrowserState, moveSelectedTo, dispatch, uploadDispatch, authHeader, wsSend }) => {
    return (
        <ColumnBox>
            <RowBox style={{ overflow: 'hidden', justifyContent: 'center' }}>
                <Card variant="solid" style={{ height: 'max-content', top: '40vh', position: 'fixed', padding: '50px' }}>
                    <Text size='20px'>
                        This folder is empty
                    </Text>

                    {filebrowserState.folderInfo.modifiable && (
                        <Card.Section style={{ display: 'flex', flexDirection: 'row', justifyContent: 'center', paddingTop: 15 }}>
                            <FileButton onChange={(files) => { HandleUploadButton(files, filebrowserState.folderInfo.id, false, "", authHeader, uploadDispatch, wsSend) }} accept="file" multiple>
                                {(props) => {
                                    return (
                                        <ColumnBox onClick={() => { props.onClick() }} style={{ cursor: 'pointer', marginTop: '0px' }}>
                                            <IconUpload size={100} style={{ padding: "10px" }} />
                                            <Text size='20px' fw={600}>
                                                Upload
                                            </Text>
                                            <Space h={4}></Space>
                                            <Text size='12px'>Click or Drop</Text>
                                        </ColumnBox>
                                    )
                                }}
                            </FileButton>
                            <Divider orientation='vertical' m={20} />

                            <ColumnBox onClick={(e) => { e.stopPropagation(); dispatch({ type: 'new_dir' }) }} style={{ cursor: 'pointer' }}>
                                <IconFolderPlus size={100} style={{ padding: "10px" }} />
                                <Text size='20px' fw={600}>
                                    New Folder
                                </Text>
                            </ColumnBox>
                        </Card.Section>
                    )}
                </Card>
            </RowBox>
        </ColumnBox>
    )
}