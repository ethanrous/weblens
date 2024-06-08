import React, { memo, useCallback, useContext, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import WeblensLoader from './Loading'

import { UserContext } from '../Context'
import { Input, Menu, Text } from '@mantine/core'
import {
    IconAlbum,
    IconExclamationCircle,
    IconFolder,
    IconLibraryPhoto,
    IconLogin,
    IconLogout,
    IconServerCog,
    IconSettings,
    IconUser,
    IconX,
} from '@tabler/icons-react'
import { WeblensButton } from './WeblensButton'
import { useKeyDown, useResize } from './hooks'
import { UpdatePassword } from '../api/UserApi'
import { AuthHeaderT, UserContextT, UserInfoT } from '../types/Types'
import Admin from '../Pages/Admin Settings/Admin'
import '../components/style.scss'

type HeaderBarProps = {
    dispatch: React.Dispatch<any>
    page: string
    loading: string[]
}

const SettingsMenu = ({
    open,
    setClosed,
    usr,
    authHeader,
}: {
    open: boolean
    setClosed: () => void
    usr: UserInfoT
    authHeader: AuthHeaderT
}) => {
    const [oldP, setOldP] = useState('')
    const [newP, setNewP] = useState('')
    useKeyDown('Escape', (e) => {
        setNewP('')
        setOldP('')
        setClosed()
    })

    const updateFunc = useCallback(async () => {
        const res =
            (await UpdatePassword(usr.username, oldP, newP, authHeader))
                .status === 200
        if (res) {
            setOldP('')
            setNewP('')
        }
        setTimeout(() => setClosed(), 2000)
        return res
    }, [usr.username, String(oldP), String(newP), authHeader])

    if (!open) {
        return null
    }
    return (
        <div className="settings-menu-container" onClick={() => setClosed()}>
            <div className="settings-menu" onClick={(e) => e.stopPropagation()}>
                <div className="flex flex-col w-max justify-center gap-2 p-24">
                    <Text
                        size="20px"
                        fw={600}
                        style={{ padding: 7, width: '90%', textWrap: 'nowrap' }}
                    >
                        Change Password
                    </Text>
                    <Input
                        value={oldP}
                        variant="unstyled"
                        type="password"
                        placeholder="Old Password"
                        className="weblens-input-wrapper"
                        onChange={(v) => setOldP(v.target.value)}
                    />
                    <Input
                        value={newP}
                        variant="unstyled"
                        type="password"
                        placeholder="New Password"
                        className="weblens-input-wrapper"
                        onChange={(v) => setNewP(v.target.value)}
                    />
                    <WeblensButton
                        squareSize={50}
                        fillWidth
                        label="Update Password"
                        showSuccess
                        disabled={oldP == '' || newP == '' || oldP === newP}
                        onClick={updateFunc}
                    />
                </div>
                <IconX
                    className="absolute top-0 left-0 cursor-pointer m-3"
                    onClick={setClosed}
                />
            </div>
        </div>
    )
}

const HeaderBar = memo(
    ({ dispatch, loading }: HeaderBarProps) => {
        const { usr, authHeader, clear, serverInfo }: UserContextT =
            useContext(UserContext)
        const nav = useNavigate()
        const [userMenu, setUserMenu] = useState(false)
        const [settings, setSettings] = useState(false)
        const [admin, setAdmin] = useState(false)
        const [barRef, setBarRef] = useState(null)
        const barSize = useResize(barRef)
        const loc = useLocation()

        const navToLogin = useCallback(() => nav('/login'), [nav])
        const navToTimeline = useCallback(() => nav('/'), [nav])
        const navToAlbums = useCallback(() => nav('/albums'), [nav])
        const navToFiles = useCallback(() => nav('/files/home'), [nav])
        const openAdmin = useCallback(() => {
            dispatch({
                type: 'set_block_focus',
                block: true,
            })
            setAdmin(true)
        }, [dispatch, setAdmin])

        const openUserSettings = useCallback(
            () => setUserMenu(true),
            [setUserMenu]
        )

        return (
            <div ref={setBarRef} className="z-50 h-max w-screen">
                {settings && (
                    <SettingsMenu
                        open={settings}
                        usr={usr}
                        setClosed={() => {
                            setSettings(false)
                            dispatch({ type: 'set_block_focus', block: false })
                        }}
                        authHeader={authHeader}
                    />
                )}
                {admin && (
                    <div
                        className="settings-menu-container"
                        onClick={() => setAdmin(false)}
                    >
                        <Admin closeAdminMenu={() => setAdmin(false)} />
                    </div>
                )}
                <div className=" absolute float-right right-10 bottom-8 z-20">
                    <WeblensLoader loading={loading} />
                </div>
                <div className="flex flex-row items-center h-14 pt-2 pb-2 border-b-2 border-neutral-700">
                    <div className="flex flex-row items-center w-96 shrink">
                        <div className="p-1" />
                        {!usr.isLoggedIn && (
                            <WeblensButton
                                label="Login"
                                squareSize={40}
                                width={barSize.width > 500 ? 112 : 40}
                                textMin={70}
                                centerContent
                                Left={IconUser}
                                onClick={navToLogin}
                            />
                        )}
                        {usr.isLoggedIn && (
                            <div className="flex flex-row items-center w-[500px]">
                                <WeblensButton
                                    label="Timeline"
                                    squareSize={40}
                                    width={barSize.width > 500 ? 112 : 40}
                                    textMin={70}
                                    centerContent
                                    subtle
                                    toggleOn={loc.pathname === '/'}
                                    Left={IconLibraryPhoto}
                                    onClick={navToTimeline}
                                />
                                <div className="p-1" />
                                <WeblensButton
                                    label="Albums"
                                    squareSize={40}
                                    width={barSize.width > 500 ? 110 : 40}
                                    textMin={60}
                                    centerContent
                                    subtle
                                    toggleOn={loc.pathname.startsWith(
                                        '/albums'
                                    )}
                                    Left={IconAlbum}
                                    onClick={navToAlbums}
                                />
                                <div className="p-1" />
                                <WeblensButton
                                    label="Files"
                                    squareSize={40}
                                    width={barSize.width > 500 ? 84 : 40}
                                    textMin={50}
                                    centerContent
                                    subtle
                                    toggleOn={loc.pathname.startsWith('/files')}
                                    Left={IconFolder}
                                    onClick={navToFiles}
                                />
                            </div>
                        )}
                    </div>
                    <div className="flex grow" />

                    {serverInfo && (
                        <div className="flex flex-col items-center h-max w-max pr-3">
                            <p className="text-xs select-none font-bold">
                                {serverInfo.name}
                            </p>
                            <p className="text-xs select-none">
                                ({serverInfo.role})
                            </p>
                        </div>
                    )}
                    <WeblensButton
                        squareSize={40}
                        width={40}
                        Left={IconExclamationCircle}
                        onClick={() => {
                            window.open(
                                `https://github.com/ethanrous/weblens/issues/new?title=Issue%20with%20${
                                    import.meta.env.VITE_APP_BUILD_TAG
                                        ? import.meta.env.VITE_APP_BUILD_TAG
                                        : 'local'
                                }`,
                                '_blank'
                            )
                        }}
                    />
                    {usr?.admin && (
                        <WeblensButton
                            squareSize={40}
                            width={40}
                            label={'Admin Settings'}
                            Left={IconServerCog}
                            onClick={openAdmin}
                        />
                    )}
                    <WeblensButton
                        squareSize={40}
                        label={'User Settings'}
                        labelOnHover
                        Left={IconUser}
                        onClick={openUserSettings}
                    />
                    <Menu opened={userMenu} onClose={() => setUserMenu(false)}>
                        <Menu.Dropdown>
                            <Menu.Label>
                                {usr.username ? usr.username : 'Not logged in'}
                            </Menu.Label>
                            <div
                                className="menu-item"
                                data-disabled={usr.username === ''}
                                onClick={() => {
                                    setUserMenu(false)
                                    setSettings(true)
                                    dispatch({
                                        type: 'set_block_focus',
                                        block: true,
                                    })
                                }}
                            >
                                <IconSettings
                                    color="white"
                                    size={20}
                                    className="cursor-pointer shrink-0"
                                />
                                <Text className="menu-item-text">Settings</Text>
                            </div>
                            {usr.username === '' && (
                                <div
                                    className="menu-item"
                                    onClick={() => {
                                        nav('/login', {
                                            state: { doLogin: false },
                                        })
                                    }}
                                >
                                    <IconLogin
                                        color="white"
                                        size={20}
                                        style={{ flexShrink: 0 }}
                                    />
                                    <Text className="menu-item-text">
                                        Login
                                    </Text>
                                </div>
                            )}
                            {usr.username !== '' && (
                                <div
                                    className="menu-item"
                                    onClick={() => {
                                        clear() // Clears cred cookies from browser
                                        nav('/login', {
                                            state: { doLogin: false },
                                        })
                                    }}
                                >
                                    <IconLogout
                                        color="white"
                                        size={20}
                                        style={{ flexShrink: 0 }}
                                    />
                                    <Text className="menu-item-text">
                                        Logout
                                    </Text>
                                </div>
                            )}
                        </Menu.Dropdown>
                    </Menu>

                    <div className="pr-3" />
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.loading !== next.loading) {
            return false
        } else if (prev.page !== next.page) {
            return false
        }
        return true
    }
)

export default HeaderBar
