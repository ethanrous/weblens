
import { memo } from 'react'

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';

import IconButton from '@mui/material/IconButton';
import UploadIcon from '@mui/icons-material/Upload'
import SyncIcon from '@mui/icons-material/Sync';
import Box from '@mui/material/Box';
import FolderIcon from '@mui/icons-material/Folder'
import PhotoLibraryIcon from '@mui/icons-material/PhotoLibrary';

import { useParams } from 'react-router-dom'

import HandleFileUpload from './Upload'
import { SendMessage } from 'react-use-websocket';

const syncDatabase = (path, sendMessage) => {
    sendMessage(JSON.stringify({
        type: 'scan_directory',
        content: {
            path: path
        },
    }))
}

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
                        <IconButton href={"/files/"} edge="start" color="inherit" aria-label="files" style={{ marginRight: 15, flexDirection: "column", fontSize: 20, minWidth: "80px" }}>
                            <FolderIcon />
                            Files
                        </IconButton>
                    )}
                    {page == "files" && (
                        <IconButton href={"/"} edge="start" color="inherit" aria-label="files" style={{ marginRight: 15, flexDirection: "column", fontSize: 20, minWidth: "80px" }}>
                            <PhotoLibraryIcon />
                            Gallery
                        </IconButton>
                    )}
                    <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); syncDatabase(path, sendMessage) }} edge="start" color="inherit" aria-label="upload" style={{ marginRight: 15, flexDirection: "column", fontSize: 20 }}>
                        <SyncIcon />
                        Sync
                    </IconButton>
                    <IconButton edge="start" color="inherit" aria-label="upload" style={{ marginRight: 15, flexDirection: "column", fontSize: 20 }}>
                        <input
                            id="upload-image"
                            hidden
                            accept="image/*"
                            type="file"
                            onChange={(e) => HandleFileUpload(e.target.files[0], "/", null)}
                        />
                        <UploadIcon />
                        Upload
                    </IconButton>
                </Toolbar>
            </AppBar>
        </Box>
    )
})

export default HeaderBar