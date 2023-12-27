
import { Ref, useContext } from 'react'
import { Search } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { dispatchSync } from '../api/Websocket'
import WeblensLoader from './Loading'

import { SendMessage } from 'react-use-websocket'
import { Sheet, Tooltip, IconButton } from '@mui/joy'

import { userContext } from '../Context'
import { Box, Input, Space } from '@mantine/core'
import { IconFolder, IconLogout, IconPhoto, IconRefresh, IconSearch, IconTools } from '@tabler/icons-react'

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
    return (
        <Box
            {...props}
            style={{
                display: 'flex',
                alignItems: 'center',
                position: 'relative',
                borderRadius: "8px",
                height: 'max-content',
                width: '100%',
                margin: '8px',
            }}
        />
    )
}

const HeaderBar = ({ folderId, searchContent, dispatch, wsSend, page, searchRef, loading, progress }: HeaderBarProps) => {
    const { userInfo, clear } = useContext(userContext)
    const nav = useNavigate()
    const spacing = "8px"

    return (
        <Box style={{ zIndex: 3, height: 'max-content', width: '100vw', position: 'fixed' }}>
            <Sheet
                sx={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    backgroundColor: "#222222",
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    height: "70px"
                }}
            >
                <Box style={{ paddingLeft: '10px' }} />
                {page === "gallery" && (
                    <Tooltip title={"Files"} disableInteractive >
                        <IconButton
                            onClick={() => nav("/files/")}
                            aria-label="files"
                            sx={{ margin: spacing }}
                        >
                            <IconFolder color='white' size={40} />
                        </IconButton>
                    </Tooltip>
                )}
                {(page === "files" || page === "admin") && (
                    <Tooltip title={"Gallery"} disableInteractive >
                        <IconButton
                            onClick={() => nav("/")}
                            aria-label="files"
                            style={{ flexDirection: "column", fontSize: 20, margin: spacing }}>
                            <IconPhoto color='white' size={40} />
                        </IconButton>
                    </Tooltip>
                )}
                <Tooltip title={"Sync"} disableInteractive >
                    <IconButton onClick={() => { dispatch({ type: 'set_loading', loading: true }); dispatchSync(folderId === "home" ? userInfo.homeFolderId : folderId, wsSend, true) }} sx={{ margin: spacing }}>
                        <IconRefresh color='white' size={40} />
                    </IconButton>
                </Tooltip>
                <SearchBox>
                    <Input
                        value={searchContent}
                        rightSection={<IconSearch />}
                        ref={searchRef}
                        placeholder="Search…"
                        onChange={e => dispatch({ type: 'set_search', search: e.target.value })}
                    />
                </SearchBox>
                <Space style={{ flexGrow: 1 }} />
                {userInfo?.admin && (
                    <Tooltip title={"Admin Settings"} disableInteractive >
                        <IconButton onClick={() => { nav("/admin") }} sx={{ margin: spacing }}>
                            <IconTools color='white' size={40} />
                        </IconButton>
                    </Tooltip>
                )}
                <Tooltip title={"Logout"} disableInteractive >
                    <IconButton onClick={() => { clear(); nav("/login", { state: { doLogin: false } }) }} sx={{ margin: spacing }}>
                        <IconLogout color='white' size={40} />
                    </IconButton>
                </Tooltip>
                <Box style={{ paddingRight: '10px' }} />
            </Sheet>
            <WeblensLoader loading={loading} progress={progress} />
        </Box>
    )
}

export default HeaderBar