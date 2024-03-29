
import { Ref, useContext } from 'react'
import { useNavigate } from 'react-router-dom'
import WeblensLoader from './Loading'

import { userContext } from '../Context'
import { ActionIcon, Box, Input, Space, Text, Tooltip } from '@mantine/core'
import { IconFolder, IconLogout, IconPhoto, IconSearch, IconTools } from '@tabler/icons-react'
import { ColumnBox, RowBox } from '../Pages/FileBrowser/FilebrowserStyles'

type HeaderBarProps = {
    searchContent: string
    dispatch: React.Dispatch<any>
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

const HeaderBar = ({ searchContent, dispatch, page, searchRef, loading, progress }: HeaderBarProps) => {
    const { userInfo, clear } = useContext(userContext)
    const nav = useNavigate()
    const spacing = "8px"

    return (
        <Box style={{ zIndex: 3, height: 'max-content', width: '100vw' }}>
            <Box
                style={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    backgroundColor: "#222222",

                    height: "70px"
                }}
            >
                <Box style={{ paddingLeft: '10px' }} />
                {page === "gallery" && (
                    <Tooltip label={"Files"} >
                        <ActionIcon
                            color='#00000000'
                            size={40}
                            onClick={() => nav("/files/")}
                            aria-label="files"
                            style={{ margin: spacing }}
                        >
                            <IconFolder color='white' size={40} />
                        </ActionIcon>
                    </Tooltip>
                )}
                {(page === "files" || page === "admin") && (
                    <Tooltip label={"Gallery"} >
                        <ActionIcon
                            color='#00000000'
                            size={40}
                            onClick={() => nav("/")}
                            aria-label="files"
                            style={{ flexDirection: "column", fontSize: 20, margin: spacing }}>
                            <IconPhoto color='white' size={40} />
                        </ActionIcon>
                    </Tooltip>
                )}
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
                <ColumnBox onClick={() => window.open(`https://github.com/ethanrous/weblens/issues/new?title=Issue%20with%20${process.env.REACT_APP_BUILD_TAG ? process.env.REACT_APP_BUILD_TAG : "local"}`, "_blank")}
                    style={{ cursor: 'pointer', width: 'max-content', alignItems: 'flex-end', justifyContent: 'center', margin: 15 }}
                >
                    <Text style={{ opacity: 0.20, zIndex: 0, userSelect: 'none' }}>{process.env.REACT_APP_BUILD_TAG ? process.env.REACT_APP_BUILD_TAG : "local"}</Text>
                    <Text lineClamp={1} style={{ opacity: 0.20, zIndex: 0, userSelect: 'none', width: 'max-content' }}>Report an issue</Text>
                </ColumnBox>
                {userInfo?.admin && (
                    <Tooltip label={"Admin Settings"} >
                        <ActionIcon color='#00000000' size={40} onClick={() => { nav("/admin") }} style={{ margin: spacing }}>
                            <IconTools size={40} />
                        </ActionIcon>
                    </Tooltip>
                )}
                <RowBox style={{ margin: spacing, backgroundColor: '#444444', height: 'max-content', width: 'max-content', padding: 6, borderRadius: '8px' }}>
                    <Text size='30' c={'white'} style={{ lineHeight: '10px', paddingBottom: 5 }}>{userInfo?.username}</Text>
                    <Space w={10} />
                    <Tooltip label={"Logout"} >
                        <ActionIcon variant='transparent' c={'white'} onClick={() => { clear(); nav("/login", { state: { doLogin: false } }) }}>
                            <IconLogout />
                        </ActionIcon>
                    </Tooltip>
                </RowBox>
                <Box style={{ paddingRight: '10px' }} />
            </Box>
            <WeblensLoader loading={loading} progress={progress} />
        </Box>
    )
}

export default HeaderBar