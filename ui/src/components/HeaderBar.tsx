import {
    IconAlbum,
    IconExclamationCircle,
    IconFolder,
    IconLibraryPhoto,
    IconLogout,
    IconServerCog,
    IconUser,
    IconX,
} from '@tabler/icons-react'
import React, { memo, useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { UpdatePassword } from '../api/UserApi'
import Admin from '../Pages/Admin Settings/Admin'
import '../components/style.scss'
import {
    AuthHeaderT,
    LOGIN_TOKEN_COOKIE_KEY,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '../types/Types'
import { useKeyDown } from './hooks'
import WeblensLoader from './Loading'
import WeblensButton from './WeblensButton'
import WeblensInput from './WeblensInput'
import { useSessionStore } from './UserInfo'
import { useCookies } from 'react-cookie'

type HeaderBarProps = {
    setBlockFocus: (block: boolean) => void
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
    const nav = useNavigate()
    const logout = useSessionStore((state) => state.logout)
    const cookies = useCookies([USERNAME_COOKIE_KEY, LOGIN_TOKEN_COOKIE_KEY])
    const deleteCookie = cookies[2]

    useKeyDown('Escape', () => {
        if (open) {
            setNewP('')
            setOldP('')
            setClosed()
        }
    })

    const updateFunc = useCallback(async () => {
        if (oldP == '' || newP == '' || oldP === newP) {
            return
        }
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
        return <></>
    }

    return (
        <div
            className="settings-menu-container"
            data-open={open}
            onClick={() => setClosed()}
        >
            <div className="settings-menu" onClick={(e) => e.stopPropagation()}>
                <div className="flex flex-row absolute right-0 top-0 p-2 m-3 bg-dark-paper rounded gap-1">
                    <IconUser />
                    <p>{usr.username}</p>
                </div>
                <div className="flex flex-col w-max justify-center items-center gap-2 p-24">
                    <p className="text-lg font-semibold p-2 w-max text-nowrap">
                        Change Password
                    </p>
                    <WeblensInput
                        placeholder="Old Password"
                        valueCallback={setOldP}
                        height={50}
                    />
                    <WeblensInput
                        onComplete={updateFunc}
                        placeholder="New Password"
                        valueCallback={setNewP}
                        height={50}
                        password
                    />
                    <div className="p2" />
                    <WeblensButton
                        label="Update Password"
                        squareSize={40}
                        fillWidth
                        showSuccess
                        disabled={oldP == '' || newP == '' || oldP === newP}
                        onClick={updateFunc}
                    />
                </div>
                <WeblensButton
                    label={'Logout'}
                    Left={IconLogout}
                    danger
                    centerContent
                    onClick={() => {
                        logout(deleteCookie)
                        nav('/login')
                    }}
                />
                <div className="top-0 left-0 m-3 absolute">
                    <WeblensButton Left={IconX} onClick={setClosed} />
                </div>
            </div>
        </div>
    )
}

const HeaderBar = memo(
    ({ setBlockFocus, loading }: HeaderBarProps) => {
        const user = useSessionStore((state) => state.user)
        const auth = useSessionStore((state) => state.auth)
        const nav = useNavigate()
        const [settings, setSettings] = useState(false)
        const [admin, setAdmin] = useState(false)
        const loc = useLocation()

        const navToTimeline = useCallback(() => nav('/'), [nav])
        const navToAlbums = useCallback(() => nav('/albums'), [nav])
        const navToFiles = useCallback(() => nav('/files/home'), [nav])

        const server = useSessionStore((state) => state.server)

        const openAdmin = useCallback(() => {
            if (setBlockFocus) {
                setBlockFocus(true)
            }
            setAdmin(true)
        }, [setBlockFocus, setAdmin])

        useEffect(() => {
            if (setBlockFocus) {
                if (settings) {
                    setBlockFocus(true)
                } else {
                    setBlockFocus(false)
                }
            }
        }, [setBlockFocus, settings])

        return (
            <div className="z-50 h-max w-screen">
                <SettingsMenu
                    open={settings}
                    usr={user}
                    setClosed={() => {
                        setBlockFocus(false)
                        setSettings(false)
                    }}
                    authHeader={auth}
                />

                {admin && (
                    <Admin
                        open={admin}
                        closeAdminMenu={() => setAdmin(false)}
                    />
                )}

                <div className=" absolute float-right right-10 bottom-8 z-20">
                    <WeblensLoader loading={loading} />
                </div>
                <div className="flex flex-row items-center h-14 pt-2 pb-2 border-b-2 border-neutral-700">
                    <div className="flex flex-row items-center w-96 shrink">
                        <div className="p-1" />
                        {user !== null && (
                            <div className="flex flex-row items-center w-full">
                                <WeblensButton
                                    label="Timeline"
                                    squareSize={40}
                                    textMin={70}
                                    centerContent
                                    toggleOn={loc.pathname === '/timeline'}
                                    Left={IconLibraryPhoto}
                                    onClick={navToTimeline}
                                    disabled={server.info.role === 'backup'}
                                />
                                <div className="p-1" />
                                <WeblensButton
                                    label="Albums"
                                    squareSize={40}
                                    textMin={60}
                                    centerContent
                                    toggleOn={loc.pathname.startsWith(
                                        '/albums'
                                    )}
                                    Left={IconAlbum}
                                    onClick={navToAlbums}
                                    disabled={server.info.role === 'backup'}
                                />
                                <div className="p-1" />
                                <WeblensButton
                                    label="Files"
                                    squareSize={40}
                                    textMin={50}
                                    centerContent
                                    toggleOn={loc.pathname.startsWith('/files')}
                                    Left={IconFolder}
                                    onClick={navToFiles}
                                />
                            </div>
                        )}
                    </div>
                    <div className="flex grow" />

                    {server && (
                        <div className="flex flex-col items-center h-max w-max pr-3">
                            <p className="text-xs select-none font-bold">
                                {server.info.name}
                            </p>
                            <p className="text-xs select-none">
                                ({server.info.role})
                            </p>
                        </div>
                    )}
                    <WeblensButton
                        labelOnHover
                        label={'Report Issue'}
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
                    {user?.admin && loc.pathname.startsWith('/files') && (
                        <WeblensButton
                            label={'Admin Settings'}
                            labelOnHover
                            Left={IconServerCog}
                            onClick={openAdmin}
                        />
                    )}
                    <WeblensButton
                        label={user !== null ? 'User Settings' : 'Login'}
                        labelOnHover
                        Left={IconUser}
                        onClick={() => {
                            if (user !== null) {
                                setSettings(true)
                            } else {
                                nav('/login')
                            }
                        }}
                    />

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
