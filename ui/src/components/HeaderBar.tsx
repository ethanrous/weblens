
import { memo } from 'react'

import AppBar from '@mui/material/AppBar'
import Toolbar from '@mui/material/Toolbar'

import IconButton from '@mui/material/IconButton'
import UploadIcon from '@mui/icons-material/Upload'
import SyncIcon from '@mui/icons-material/Sync'
import LogoutIcon from '@mui/icons-material/Logout'
import Box from '@mui/material/Box'
import FolderIcon from '@mui/icons-material/Folder'
import PhotoLibraryIcon from '@mui/icons-material/PhotoLibrary'
import Tooltip from '@mui/material/Tooltip'

import { useNavigate, useParams } from 'react-router-dom'

import HandleFileUpload from '../api/Upload'
import { dispatchSync } from '../api/Websocket'

import { SendMessage } from 'react-use-websocket'
import { TextField, alpha, styled } from '@mui/material'
import InputBase from '@mui/material/InputBase'
import SearchIcon from '@mui/icons-material/Search'
import { useCookies } from 'react-cookie'

type HeaderBarProps = {
    dispatch: React.Dispatch<any>
    wsSend: SendMessage
    page: string,
}

const Search = styled('div')(({ theme }) => ({
    position: 'relative',
    borderRadius: theme.shape.borderRadius,
    backgroundColor: alpha(theme.palette.common.white, 0.15),
    '&:hover': {
        backgroundColor: alpha(theme.palette.common.white, 0.25),
    },
    marginLeft: 0,
    width: '100%',
    [theme.breakpoints.up('sm')]: {
        marginLeft: theme.spacing(1),
        width: 'auto',
    },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
    padding: theme.spacing(0, 2),
    height: '100%',
    position: 'absolute',
    pointerEvents: 'none',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
    color: 'inherit',
    '& .MuiInputBase-input': {
        padding: theme.spacing(1, 1, 1, 0),
        // vertical padding + font size from searchIcon
        paddingLeft: `calc(1em + ${theme.spacing(4)})`,
        transition: theme.transitions.create('width'),
        width: '100%',
        [theme.breakpoints.up('sm')]: {
            width: '12ch',
            '&:focus': {
                width: '20ch',
            },
        },
    },
}));

const HeaderBar = memo(function HeaderBar({ dispatch, wsSend, page }: HeaderBarProps) {
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-login-token']);
    const nav = useNavigate()

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
                            <IconButton onClick={() => nav("/files/")} edge="start" color="inherit" aria-label="files" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                                <FolderIcon />
                            </IconButton>
                        </Tooltip>
                    )}
                    {page == "files" && (
                        <Tooltip title={"Gallery"}>
                            <IconButton onClick={() => nav("/")} edge="start" color="inherit" aria-label="files" style={{ margin: 15, flexDirection: "column", fontSize: 20 }}>
                                <PhotoLibraryIcon />
                            </IconButton>
                        </Tooltip>

                    )}
                    {page == "files" && (
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
                    )}
                    <Tooltip title={"Sync"}>
                        <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(path, wsSend, true) }} edge="start" color="inherit" aria-label="upload" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                            <SyncIcon />
                        </IconButton>
                    </Tooltip>

                    <Search>
                        <SearchIconWrapper>
                            <SearchIcon />
                        </SearchIconWrapper>
                        <StyledInputBase
                            placeholder="Search…"
                            inputProps={{ 'aria-label': 'search' }}
                            onChange={e => dispatch({ type: 'set_search', search: e.target.value })}
                        />
                    </Search>
                    <Tooltip title={"Logout"} >
                        <IconButton edge="end" color="inherit" aria-label="logout" style={{ margin: 10, flexDirection: "column", fontSize: 20 }} onClick={() => { removeCookie('weblens-login-token'); nav("/login") }}>
                            <LogoutIcon />
                        </IconButton>
                    </Tooltip>
                </Toolbar>
            </AppBar>
        </Box>
    )
})

export default HeaderBar