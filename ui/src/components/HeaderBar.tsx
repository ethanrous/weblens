import {
    IconFolder,
    IconLibraryPhoto,
    IconMoon,
    IconSettings,
    IconSun,
} from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ThemeStateEnum, useWlTheme } from '@weblens/store/ThemeControl'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { useCallback } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import { useSessionStore } from './UserInfo'
import headerBarStyle from './headerBarStyle.module.scss'

export const ThemeToggleButton = () => {
    const { theme, isOSControlled, changeTheme } = useWlTheme()

    return (
        <>
            <WeblensButton
                label="Theme"
                disabled={isOSControlled}
                Left={theme === ThemeStateEnum.DARK ? IconMoon : IconSun}
                onClick={() => {
                    changeTheme(
                        theme === ThemeStateEnum.DARK
                            ? ThemeStateEnum.LIGHT
                            : ThemeStateEnum.DARK
                    )
                }}
            />
            <WeblensButton
                label="Os Theme"
                centerContent
                toggleOn={isOSControlled}
                allowRepeat
                onClick={() => {
                    changeTheme(
                        isOSControlled ? ThemeStateEnum.DARK : ThemeStateEnum.OS
                    )
                }}
            />
        </>
    )
}

function HeaderBar() {
    const { user } = useSessionStore()
    const clearMedia = useMediaStore((state) => state.clear)
    const clearFiles = useFileBrowserStore((state) => state.clearFiles)
    const nav = useNavigate()
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
            <div className={headerBarStyle.headerBar}>
                <div className={headerBarStyle.navBox}>
                    <div className="p-1" />
                    {user !== null && (
                        <div className="flex w-[140px] grow flex-row items-center gap-2">
                            <WeblensButton
                                label="Files"
                                flavor={
                                    loc.pathname.startsWith('/files')
                                        ? 'default'
                                        : 'outline'
                                }
                                Left={IconFolder}
                                onClick={navToFiles}
                            />
                            <WeblensButton
                                label="Timeline"
                                squareSize={36}
                                textMin={70}
                                centerContent
                                flavor={
                                    loc.pathname.startsWith('/timeline')
                                        ? 'default'
                                        : 'outline'
                                }
                                Left={IconLibraryPhoto}
                                onClick={navToTimeline}
                                disabled={
                                    server.role === 'backup' || !user.isLoggedIn
                                }
                            />
                        </div>
                    )}
                </div>
                <div className="ml-auto">
                    <WeblensButton
                        label={!user.isLoggedIn ? 'Login' : ''}
                        tooltip={!user.isLoggedIn ? 'Login' : 'Settings'}
                        Left={IconSettings}
                        disabled={window.location.pathname.startsWith(
                            '/settings'
                        )}
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
                </div>
                <div className="pr-4" />
            </div>
        </div>
    )
}

export default HeaderBar
