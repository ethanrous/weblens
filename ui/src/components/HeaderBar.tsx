
import { Ref, memo, useRef } from 'react'

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
import WeblensLoader from './Loading'

import { SendMessage } from 'react-use-websocket'
import { Paper, alpha, styled } from '@mui/material'
import InputBase from '@mui/material/InputBase'
import SearchIcon from '@mui/icons-material/Search'
import { useCookies } from 'react-cookie'

type HeaderBarProps = {
    dispatch: React.Dispatch<any>
    wsSend: SendMessage
    page: string
    searchRef: Ref<any>
    loading: boolean
    progress: number
}

const Search = styled('div')(({ theme }) => ({
    position: 'relative',
    borderRadius: theme.shape.borderRadius,
    backgroundColor: alpha(theme.palette.common.white, 0.15),
    '&:hover': {
        backgroundColor: alpha(theme.palette.common.white, 0.25),
    },
    width: '100%',
    [theme.breakpoints.up('sm')]: {

        width: 'auto',
    },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
    padding: theme.spacing(0, 1),
    height: '100%',
    position: 'absolute',
    pointerEvents: 'none',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
    '& .MuiInputBase-input': {
        padding: theme.spacing(8, 2, 8, 0),
        paddingLeft: `calc(1em + ${theme.spacing(20)})`,
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

const HeaderBar = memo(function HeaderBar({ dispatch, wsSend, page, searchRef, loading, progress }: HeaderBarProps) {
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token'])
    const nav = useNavigate()

    let path = (useParams()["*"] + "/").replace(/\/\/+/g, '/')
    if (page == "gallery") {
        path = "/"
    }

    return (
        <Box zIndex={1} height={"max-content"} width={"100vw"} position={'fixed'} >
            <Paper sx={{ width: "100%", backgroundColor: (theme) => { return alpha(theme.palette.secondary.dark, 0.80) }, backdropFilter: "blur(8px)", border: (theme) => `1px solid ${theme.palette.divider}`, height: "75px" }}>
                <Toolbar style={{ paddingLeft: "25px", width: "100%" }}>
                    {page == "gallery" && (
                        <Tooltip title={"Files"} disableInteractive >
                            <IconButton onClick={() => nav("/files/")} edge="start" color="primary" aria-label="files" style={{ margin: 10 }}>
                                <FolderIcon fontSize={'large'} />
                            </IconButton>
                        </Tooltip>
                    )}
                    {page == "files" && (
                        <Tooltip title={"Gallery"} disableInteractive >
                            <IconButton onClick={() => nav("/")} edge="start" color="primary" aria-label="files" style={{ margin: 10, flexDirection: "column", fontSize: 20 }}>
                                <PhotoLibraryIcon fontSize={'large'} />
                            </IconButton>
                        </Tooltip>

                    )}
                    {page == "files" && (
                        <Tooltip title={"Upload"} disableInteractive >
                            <IconButton edge="start" color="primary" aria-label="upload" style={{ margin: 10 }}>
                        <input
                            id="upload-image"
                            hidden
                            accept="image/*"
                            type="file"
                            onChange={(e) => HandleFileUpload(e.target.files[0], "/", null)}
                            />
                                <UploadIcon fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                    )}
                    <Tooltip title={"Sync"} disableInteractive >
                        <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(path, wsSend, true) }} sx={{ margin: 10 }} >
                            <SyncIcon fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                    <Search sx={{ margin: 10 }}>
                        <SearchIconWrapper>
                            <SearchIcon fontSize={'medium'} sx={{ marginLeft: "6px" }} />
                        </SearchIconWrapper>
                        <StyledInputBase
                            ref={searchRef}
                            placeholder="Search…"
                            inputProps={{ 'aria-label': 'search' }}
                            onChange={e => dispatch({ type: 'set_search', search: e.target.value })}
                        />
                    </Search>
                    <Tooltip title={"Logout"} disableInteractive >
                        <IconButton onClick={() => { removeCookie('weblens-login-token'); nav("/login") }} sx={{ margin: 10 }}>
                            <LogoutIcon fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                </Toolbar>

            </Paper>
            <WeblensLoader loading={loading} progress={progress} />
        </Box>
    )
})

export default HeaderBar