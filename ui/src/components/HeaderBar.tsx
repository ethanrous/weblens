
import { memo } from 'react'

import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'

import IconButton from '@mui/material/IconButton'
import UploadIcon from '@mui/icons-material/Upload'
import SyncIcon from '@mui/icons-material/Sync'
import Box from '@mui/material/Box'
import FolderIcon from '@mui/icons-material/Folder'
import PhotoLibraryIcon from '@mui/icons-material/PhotoLibrary'
import Tooltip from '@mui/material/Tooltip'

import { useParams } from 'react-router-dom'

import HandleFileUpload from '../api/Upload'
import { dispatchSync } from '../api/Websocket'

import { SendMessage } from 'react-use-websocket';

type HeaderBarProps = {
    dispatch: React.Dispatch<any>
    sendMessage: SendMessage
    page: string
}

const HeaderBar = memo(function HeaderBar({ dispatch, sendMessage, page }: HeaderBarProps) {
    let path = (useParams()["*"] + "/").replace(/\/\/+/g, '/')
    if (page == "gallery") {
        path = "/"
    }

    return (
        <Box sx={{ flexGrow: 1 }} zIndex={1} pt={1} maxWidth={"100vw"}>
            <AppBar
                position="static"
                color='transparent'
                style={{ boxShadow: "none" }}
            >
                <Toolbar style={{ paddingLeft: "25px" }}>
                    {page == "gallery" && (
                        <Tooltip title={"Files"}>
                            <IconButton href={"/files/"} edge="start" color="inherit" aria-label="files" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                                <FolderIcon />
                            </IconButton>
                        </Tooltip>
                    )}
                    {page == "files" && (
                        <Tooltip title={"Gallery"}>
                            <IconButton href={"/"} edge="start" color="inherit" aria-label="files" style={{ margin: 15, flexDirection: "column", fontSize: 20 }}>
                                <PhotoLibraryIcon />
                            </IconButton>
                        </Tooltip>
                    )}
                    <Tooltip title={"Sync"}>
                        <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(path, sendMessage, true) }} edge="start" color="inherit" aria-label="upload" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                            <SyncIcon />
                    </IconButton>
                    </Tooltip>
                    <Tooltip title={"Upload"}>
                        <IconButton edge="start" color="inherit" aria-label="upload" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                        <input
                            id="upload-image"
                            hidden
                            accept="image/*"
                            type="file"
                            onChange={(e) => HandleFileUpload(e.target.files[0], "/", null)}
                            />
                        <UploadIcon />

                    </IconButton>
                    </Tooltip>
                </Toolbar>
            </AppBar>
        </Box>
    )
})

export default HeaderBar