
import { Ref, useContext, useRef } from 'react'

import { Folder, Search, Sync, Upload, AdminPanelSettings, Logout } from '@mui/icons-material'
import PhotoLibraryIcon from '@mui/icons-material/PhotoLibrary'

import { useNavigate } from 'react-router-dom'

// import HandleFileUpload from '../api/Upload'
import { dispatchSync } from '../api/Websocket'
import WeblensLoader from './Loading'

import { SendMessage } from 'react-use-websocket'
import { Sheet, Typography, styled, Input, Tooltip, Box, IconButton, useTheme } from '@mui/joy'

import { userContext } from '../Context'

type HeaderBarProps = {
    folderId: string
    searchContent: string
    dispatch: React.Dispatch<any>
    wsSend: SendMessage
    page: string
    searchRef: Ref<any>
    loading: boolean
    progress: number
}

const SearchBox = ({ ...props }) => {
    let theme = useTheme()
    return (
        <Box
            {...props}
            sx={{
                display: 'flex',
                alignItems: 'center',
                position: 'relative',
                borderRadius: "8px",
                height: 'max-content',
                width: '100%',
                margin: '8px',
                [theme.breakpoints.up('sm')]: {
                    width: 'auto',
                },
            }}
        />
    )
}

const StyledInput = styled(Input)(({ theme }) => ({
    // backgroundColor: theme.colorSchemes.dark.palette.primary.softBg,
    backgroundColor: `rgba(0 0 0 / 0.0)`,
    width: '15ch',
    transition: "width 0.5s",
    '&.Mui-focused': {
        backgroundColor: theme.colorSchemes.dark.palette.neutral.softBg,
        width: '20ch',
        transition: "width 0.5s"
    },
    '& .MuiInputBase-input': {
        // color: theme.colorSchemes.dark.palette.neutral.softBg,
        padding: theme.spacing(8, 2, 8, 0),
        paddingLeft: `calc(1em + ${theme.spacing(20)})`,
        width: '100%',
        [theme.breakpoints.up('sm')]: {
            width: '12ch',
            '&:focus': {
                width: '20ch',
            },
        },
    },
}));

const HeaderBar = ({ folderId, searchContent, dispatch, wsSend, page, searchRef, loading, progress }: HeaderBarProps) => {
    const { authHeader, userInfo, clear } = useContext(userContext)
    const nav = useNavigate()
    const spacing = "8px"
    const theme = useTheme()
    const openRef = useRef(null);
    return (
        <Box zIndex={3} height={"max-content"} width={"100vw"} position={'fixed'} >
            <Sheet
                sx={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    backgroundColor: theme.colorSchemes.dark.palette.primary.softBg,
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    height: "70px"
                }}
            >
                <Box paddingLeft={'10px'} />
                {page === "gallery" && (
                    <Tooltip title={"Files"} disableInteractive >
                        <IconButton
                            onClick={() => nav("/files/")}
                            aria-label="files"
                            sx={{ margin: spacing }}
                        >
                            <Folder fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                )}
                {(page === "files" || page === "admin") && (
                    <Tooltip title={"Gallery"} disableInteractive >
                        <IconButton
                            onClick={() => nav("/")}
                            aria-label="files"
                            style={{ flexDirection: "column", fontSize: 20, margin: spacing }}>
                            <PhotoLibraryIcon fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                )}
                <Tooltip title={"Sync"} disableInteractive >
                    <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(folderId === "home" ? userInfo.homeFolderId : folderId, wsSend, true) }} sx={{ margin: spacing }}>
                        <Sync fontSize={'large'} />
                    </IconButton>
                </Tooltip>
                <SearchBox>
                    <StyledInput
                        value={searchContent}
                        endDecorator={<Search />}
                        color='neutral'
                        ref={searchRef}
                        placeholder="Search…"
                        onChange={e => dispatch({ type: 'set_search', search: e.target.value })}
                    />
                </SearchBox>
                {(loading && progress !== 0 && progress !== 100) && (
                    <Sheet sx={{
                        height: "max-content",
                        padding: "10px",
                        marginLeft: "20px",
                        borderRadius: '8px',
                        backgroundColor: (theme) => { return theme.colorSchemes.dark.palette.neutral.softBg }
                    }}>
                        <Typography color='primary'>
                            Directory Scan: {progress.toFixed(0)}%
                        </Typography>
                    </Sheet>
                )}
                <Box flexGrow={1} />
                {userInfo?.admin && (
                    <Tooltip title={"Admin Settings"} disableInteractive >
                        <IconButton onClick={() => { nav("/admin") }} sx={{ margin: spacing }}>
                            <AdminPanelSettings fontSize={'large'} />
                        </IconButton>
                    </Tooltip>
                )}
                <Tooltip title={"Logout"} disableInteractive >
                    <IconButton onClick={() => { clear(); nav("/login", { state: { doLogin: false } }) }} sx={{ margin: spacing }}>
                        <Logout fontSize={'large'} />
                    </IconButton>
                </Tooltip>
                <Box paddingRight={'10px'} />
            </Sheet>
            <WeblensLoader loading={loading} progress={progress} />
        </Box>
    )
}

export default HeaderBar