import {
    IconFolder,
    IconLibraryPhoto,
    IconMoon,
    IconServerCog,
    IconSun,
    IconUser,
} from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton'
import Admin from '@weblens/pages/AdminSettings/Admin'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { toggleLightTheme } from '@weblens/util'
import { useCallback, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { useSessionStore } from './UserInfo'
import headerBarStyle from './headerBarStyle.module.scss'

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

function HeaderBar() {
    const { user } = useSessionStore()
    const clearMedia = useMediaStore((state) => state.clear)
    const clearFiles = useFileBrowserStore((state) => state.clearFiles)
    const nav = useNavigate()
    const [admin, setAdmin] = useState(false)
    const loc = useLocation()

    const navToTimeline = useCallback(() => {
        clearFiles()
        clearMedia()
        nav('/timeline')
    }, [nav])

    const navToFiles = useCallback(() => {
        clearFiles()
        clearMedia()
        nav('/files/home')
    }, [nav])

    const server = useSessionStore((state) => state.server)

    return (
        <div className="z-50 h-max w-screen">
            {(admin || server.role === 'backup') && (
                <Admin closeAdminMenu={() => setAdmin(false)} />
            )}

            <div className={headerBarStyle['header-bar']}>
                <div className={headerBarStyle['nav-box']}>
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
                                toggleOn={loc.pathname.startsWith('/timeline')}
                                Left={IconLibraryPhoto}
                                onClick={navToTimeline}
                                disabled={
                                    server.role === 'backup' || !user.isLoggedIn
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

                {user?.admin && (
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
                    disabled={window.location.pathname.startsWith('/settings')}
                    onClick={() => {
                        if (user.isLoggedIn) {
                            nav('/settings')
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
}

export default HeaderBar
