import {
    IconAlbum,
    IconFolder,
    IconLibraryPhoto,
    IconLogout,
    IconMoon,
    IconServer,
    IconServerCog,
    IconSun,
    IconUser,
    IconX,
} from '@tabler/icons-react'
import { UpdatePassword } from '@weblens/api/UserApi'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import Admin from '@weblens/pages/Admin Settings/Admin'
import '@weblens/components/style.scss'
import '@weblens/components/theme.scss'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '@weblens/types/Types'
import { memo, useCallback, useEffect, useState } from 'react'
import { useCookies } from 'react-cookie'
import { useLocation, useNavigate } from 'react-router-dom'
import { useKeyDown } from './hooks'
import WeblensLoader from './Loading'
import { useSessionStore } from './UserInfo'

type HeaderBarProps = {
    setBlockFocus: (block: boolean) => void
    page: string
    loading: string[]
}

export const ThemeToggleButton = () => {
    const [isDarkTheme, setIsDarkTheme] = useState(
        document.documentElement.classList.contains('dark')
    )
    return (
        <WeblensButton
            label="Theme"
            Left={isDarkTheme ? IconMoon : IconSun}
            onClick={() => {
                localStorage.setItem('theme', isDarkTheme ? 'light' : 'dark')
                document.documentElement.classList.toggle('dark')
                setIsDarkTheme(!isDarkTheme)
            }}
        />
    )
}

const SettingsMenu = ({
    open,
    setClosed,
    user,
}: {
    open: boolean
    setClosed: () => void
    user: UserInfoT
}) => {
    const [oldP, setOldP] = useState('')
    const [newP, setNewP] = useState('')
    const [buttonRef, setButtonRef] = useState<HTMLDivElement>()
    const nav = useNavigate()
    const logout = useSessionStore((state) => state.logout)
    const [, , deleteCookie] = useCookies([
        USERNAME_COOKIE_KEY,
        LOGIN_TOKEN_COOKIE_KEY,
    ])

    useKeyDown('Escape', () => {
        if (open) {
            setNewP('')
            setOldP('')
            setClosed()
        }
    })

    const updateFunc = useCallback(async () => {
        if (oldP == '' || newP == '' || oldP === newP) {
            return Promise.reject(
                'Old and new password cannot be empty or match'
            )
        }
        return UpdatePassword(user.username, oldP, newP).then((r) => {
            if (r.status !== 200) {
                return Promise.reject(r.statusText)
            }
            setTimeout(() => {
                setNewP('')
                setOldP('')
            }, 2000)
            return true
        })
    }, [user.username, String(oldP), String(newP)])

    useKeyDown('Enter', () => {
        if (buttonRef) {
            buttonRef.click()
        }
    })

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
                <div className="theme-dark-paper flex flex-row absolute right-0 top-0 p-2 m-3 rounded gap-1">
                    <IconUser className="theme-text-dark-bg" />
                    <p className="theme-text-dark-bg">{user.username}</p>
                </div>
                <ThemeToggleButton />
                <div className="flex flex-col w-max justify-center items-center gap-2 p-24">
                    <p className="text-lg font-semibold p-2 w-max text-nowrap">
                        Change Password
                    </p>
                    <WeblensInput
                        value={oldP}
                        placeholder="Old Password"
                        password
                        valueCallback={setOldP}
                        squareSize={50}
                    />
                    <WeblensInput
                        value={newP}
                        placeholder="New Password"
                        valueCallback={setNewP}
                        squareSize={50}
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
                        setButtonRef={setButtonRef}
                    />
                </div>
                <WeblensButton
                    label={'Logout'}
                    Left={IconLogout}
                    danger
                    centerContent
                    onClick={() => {
                        useMediaStore.getState().clear()
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
        const { user } = useSessionStore()
        const clearMedia = useMediaStore((state) => state.clear)
        const nav = useNavigate()
        const [settings, setSettings] = useState(false)
        const [admin, setAdmin] = useState(false)
        const loc = useLocation()

        const navToTimeline = useCallback(() => {
            clearMedia()
            nav('/')
        }, [nav])
        const navToAlbums = useCallback(() => {
            clearMedia()
            nav('/albums')
        }, [nav])
        const navToFiles = useCallback(() => {
            clearMedia()
            nav('/files/home')
        }, [nav])

        const server = useSessionStore((state) => state.server)

        useEffect(() => {
            if (settings || admin) {
                setBlockFocus(true)
            } else if (!settings && !admin) {
                setBlockFocus(false)
            }
        }, [settings, admin])

        return (
            <div className="z-50 h-max w-screen">
                {settings && (
                    <SettingsMenu
                        open={settings}
                        user={user}
                        setClosed={() => {
                            setSettings(false)
                        }}
                    />
                )}

                {(admin || server.info.ole === 'backup') && (
                    <Admin closeAdminMenu={() => setAdmin(false)} />
                )}

                <div className=" absolute float-right right-10 bottom-8 z-20">
                    <WeblensLoader loading={loading} />
                </div>
                <div className="header-bar">
                    <div className="flex flex-row items-center w-96 shrink">
                        <div className="p-1" />
                        {user !== null && (
                            <div className="flex flex-row items-center w-[140px] grow">
                                <WeblensButton
                                    label="Timeline"
                                    squareSize={36}
                                    textMin={70}
                                    centerContent
                                    toggleOn={loc.pathname === '/timeline'}
                                    Left={IconLibraryPhoto}
                                    onClick={navToTimeline}
                                    disabled={
                                        server.info.role === 'backup' ||
                                        !user.isLoggedIn
                                    }
                                />
                                <WeblensButton
                                    label="Albums"
                                    squareSize={36}
                                    textMin={60}
                                    centerContent
                                    toggleOn={loc.pathname.startsWith(
                                        '/albums'
                                    )}
                                    Left={IconAlbum}
                                    onClick={navToAlbums}
                                    disabled={
                                        server.info.role === 'backup' ||
                                        !user.isLoggedIn
                                    }
                                />
                                <WeblensButton
                                    label="Files"
                                    squareSize={36}
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
                        <div
                            className="flex bg-dark-paper items-center justify-center h-max w-max p-2 rounded gap-1 cursor-pointer"
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
                        >
                            <IconServer
                                size={20}
                                className="theme-text-dark-bg"
                            />
                            <p className="theme-text-dark-bg text-xs select-none font-bold">
                                {server.info.name}
                            </p>
                            <p className="theme-text-dark-bg text-xs select-none">
                                {server.info.role}
                            </p>
                            <div
                                className="button-tooltip"
                                style={{
                                    transform: `translateY(35px)`,
                                }}
                            >
                                <p className="text-white">
                                    {import.meta.env.VITE_APP_BUILD_TAG
                                        ? import.meta.env.VITE_APP_BUILD_TAG
                                        : 'local'}
                                </p>
                            </div>
                        </div>
                    )}
                    {user?.admin && loc.pathname.startsWith('/files') && (
                        <WeblensButton
                            label={'Admin Settings'}
                            labelOnHover
                            Left={IconServerCog}
                            onClick={() => setAdmin(true)}
                        />
                    )}
                    <WeblensButton
                        label={user.isLoggedIn ? 'User Settings' : 'Login'}
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
