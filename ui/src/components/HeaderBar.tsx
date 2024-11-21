import {
    IconFolder,
    IconLibraryPhoto,
    IconLogout,
    IconMoon,
    IconServerCog,
    IconSun,
    IconUser,
    IconX,
} from '@tabler/icons-react'
import UsersApi from '@weblens/api/UserApi'
import { style } from '@weblens/components/style'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import Admin from '@weblens/pages/Admin Settings/Admin'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import User from '@weblens/types/user/User'
import { toggleLightTheme } from '@weblens/util'
import { memo, useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { useSessionStore } from './UserInfo'
import { useKeyDown } from './hooks'

type HeaderBarProps = {
    setBlockFocus: (block: boolean) => void
    page: string
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
                setIsDarkTheme(toggleLightTheme())
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
    user: User
}) => {
    const [oldP, setOldP] = useState('')
    const [newP, setNewP] = useState('')
    const [buttonRef, setButtonRef] = useState<HTMLDivElement>()
    const setUser = useSessionStore((state) => state.setUser)
    const nav = useNavigate()

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
                new Error('Old and new password cannot be empty or match')
            )
        }
        return UsersApi.updateUserPassword(user.username, {
            oldPassword: oldP,
            newPassword: newP,
        }).then(() => {
            setNewP('')
            setOldP('')
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
            className={style['settings-menu-container']}
            data-open={open}
            onClick={() => setClosed()}
        >
            <div
                className={style['settings-menu']}
                onClick={(e) => e.stopPropagation()}
            >
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
                    onClick={async () => {
                        useMediaStore.getState().clear()
                        await UsersApi.logoutUser()
                        const loggedOut = new User()
                        loggedOut.isLoggedIn = false
                        setUser(loggedOut)
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
    ({ setBlockFocus }: HeaderBarProps) => {
        const { user } = useSessionStore()
        const clearMedia = useMediaStore((state) => state.clear)
        const clearFiles = useFileBrowserStore((state) => state.clearFiles)
        const nav = useNavigate()
        const [settings, setSettings] = useState(false)
        const [admin, setAdmin] = useState(false)
        const loc = useLocation()

        const navToTimeline = useCallback(() => {
            clearFiles()
            clearMedia()
            nav('/timeline')
        }, [nav])
        // const navToAlbums = useCallback(() => {
        //     clearFiles()
        //     clearMedia()
        //     nav('/albums')
        // }, [nav])
        const navToFiles = useCallback(() => {
            clearFiles()
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

                {(admin || server.role === 'backup') && (
                    <Admin closeAdminMenu={() => setAdmin(false)} />
                )}

                <div className={style['header-bar']}>
                    <div className={style['nav-box']}>
                        <div className="p-1" />
                        {user !== null && (
                            <div className="flex flex-row items-center w-[140px] grow">
                                <WeblensButton
                                    label="Files"
                                    squareSize={36}
                                    textMin={60}
                                    centerContent
                                    toggleOn={loc.pathname.startsWith('/files')}
                                    Left={IconFolder}
                                    onClick={navToFiles}
                                />
                                <WeblensButton
                                    label="Timeline"
                                    squareSize={36}
                                    textMin={70}
                                    centerContent
                                    toggleOn={loc.pathname.startsWith(
                                        '/timeline'
                                    )}
                                    Left={IconLibraryPhoto}
                                    onClick={navToTimeline}
                                    disabled={
                                        server.role === 'backup' ||
                                        !user.isLoggedIn
                                    }
                                />
                                {/* <WeblensButton */}
                                {/*     label="Albums" */}
                                {/*     squareSize={36} */}
                                {/*     textMin={60} */}
                                {/*     centerContent */}
                                {/*     toggleOn={loc.pathname.startsWith( */}
                                {/*         '/albums' */}
                                {/*     )} */}
                                {/*     Left={IconAlbum} */}
                                {/*     onClick={navToAlbums} */}
                                {/*     disabled={ */}
                                {/*         server.role === 'backup' || */}
                                {/*         !user.isLoggedIn */}
                                {/*     } */}
                                {/* /> */}
                            </div>
                        )}
                    </div>
                    <div className="flex grow" />

                    {user?.admin && loc.pathname.startsWith('/files') && (
                        <WeblensButton
                            Left={IconServerCog}
                            tooltip="Server Settings"
                            onClick={() => setAdmin(true)}
                        />
                    )}
                    <WeblensButton
                        label={!user.isLoggedIn ? 'Login' : ''}
                        tooltip={!user.isLoggedIn ? 'Login' : 'Me'}
                        Left={IconUser}
                        onClick={() => {
                            if (user.isLoggedIn) {
                                setSettings(true)
                            } else {
                                nav('/login', {
                                    state: {
                                        returnTo: window.location.pathname,
                                    },
                                })
                            }
                        }}
                    />

                    <div className="pr-3" />
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.page !== next.page) {
            return false
        }
        return true
    }
)

export default HeaderBar
