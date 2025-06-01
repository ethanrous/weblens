import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MediaApi from '@weblens/api/MediaApi'
import ErrorBoundary from '@weblens/components/Error.tsx'
import Logo from '@weblens/components/Logo.tsx'
import Messages from '@weblens/components/Messages.tsx'
import useR, { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown } from '@weblens/lib/hooks'
import Fourohfour from '@weblens/pages/404/fourohfour'
import Backup from '@weblens/pages/Backup/Backup'
import Signup from '@weblens/pages/Signup/Signup'
import StartUp from '@weblens/pages/Startup/StartupPage'
import { ThemeStateEnum, useWlTheme } from '@weblens/store/ThemeControl'
import { ErrorHandler } from '@weblens/types/Types'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import axios from 'axios'
import React, { Suspense, useCallback, useEffect } from 'react'
import {
    BrowserRouter as Router,
    useLocation,
    useNavigate,
    useRoutes,
} from 'react-router-dom'

import WeblensLoader from './components/Loading'

const Gallery = React.lazy(() => import('@weblens/pages/Gallery/Gallery'))
const FileBrowser = React.lazy(
    () => import('@weblens/pages/FileBrowser/FileBrowser')
)
const Login = React.lazy(() => import('@weblens/pages/Login/Login'))
const Setup = React.lazy(() => import('@weblens/pages/Setup/Setup'))
const SettingsMenu = React.lazy(
    () => import('@weblens/pages/UserSettings/Settings')
)

axios.defaults.withCredentials = true

const WeblensRoutes = () => {
    useR()

    const fetchServerInfo = useSessionStore((state) => state.fetchServerInfo)
    const server = useSessionStore((state) => state.server)
    const serverFetchError = useSessionStore((state) => state.serverFetchError)
    const user = useSessionStore((state) => state.user)

    const setNav = useSessionStore((state) => state.setNav)

    const loc = useLocation()
    const nav = useNavigate()

    const setTypeMap = useMediaStore((state) => state.setTypeMap)

    useEffect(() => {
        fetchServerInfo().catch(ErrorHandler)
    }, [fetchServerInfo])

    useEffect(() => {
        if (nav) {
            setNav(nav)
        }
    }, [nav, setNav])

    useEffect(() => {
        if (!server) {
            return
        } else if (loc.pathname !== '/setup' && server.role === 'init') {
            console.debug('Nav setup')
            nav('/setup')
        } else if (loc.pathname === '/setup' && server.role === 'core') {
            console.debug('Nav files home from setup')
            nav('/files/home')
        } else if (
            server.role === 'backup' &&
            loc.pathname !== '/backup' &&
            !loc.pathname.startsWith('/settings') &&
            user?.isLoggedIn
        ) {
            console.debug('Nav backup page from', loc.pathname)
            nav('/backup')
        } else if (loc.pathname === '/login' && user?.isLoggedIn) {
            const state = loc.state as {
                returnTo: string
            }
            if (state?.returnTo !== undefined) {
                console.debug('Nav return to', state.returnTo)
                nav(state.returnTo)
            } else {
                console.debug('Nav files home from login')
                nav('/files/home')
            }
        } else if (loc.pathname === '/' && server.role === 'core') {
            console.debug('Nav files home (root path)')
            nav('/files/home')
        }
    }, [loc, server, user, nav])

    useEffect(() => {
        if (!server || !server.started || server.role === 'init') {
            return
        }
        MediaApi.getMediaTypes()
            .then((r) => {
                setTypeMap(r.data)
            })
            .catch(ErrorHandler)
    }, [server, setTypeMap])

    useEffect(() => {
        const theme = localStorage.getItem('theme')
        if (theme === 'dark') {
            document.documentElement.style.setProperty(
                'color-scheme',
                ThemeStateEnum.DARK
            )
        }
    }, [])

    if (serverFetchError) {
        return (
            <FatalError
                err={'Failed to get server info. Is the server running?'}
            />
        )
    }

    if (!server || !user) {
        return <WeblensLoader className="m-auto" />
    }

    return (
        <ErrorBoundary>
            {server.started && <PageSwitcher />}
            {!server.started && <StartUp />}
        </ErrorBoundary>
    )
}

function FatalError({ err }: { err: string }) {
    return (
        <div className="flex h-screen w-screen flex-col items-center justify-center gap-6">
            <h2 className="text-red-500">Fatal Error</h2>
            <h3>{err}</h3>

            <Logo />
        </div>
    )
}

function PageSwitcher() {
    const user = useSessionStore((state) => state.user)
    const loc = useLocation()

    const galleryPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Gallery />
        </Suspense>
    )

    const filesPage = (user.isLoggedIn ||
        loc.pathname.startsWith('/files/share')) && (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <FileBrowser />
        </Suspense>
    )

    const loginPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Login />
        </Suspense>
    )

    const signupPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Signup />
        </Suspense>
    )

    const setupPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Setup />
        </Suspense>
    )

    const backupPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Backup />
        </Suspense>
    )

    const settingsPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <SettingsMenu />
        </Suspense>
    )

    const fourOhFourPage = (
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <Fourohfour />
        </Suspense>
    )

    const Gal = useRoutes([
        { path: '/', element: galleryPage },
        { path: '/timeline', element: galleryPage },
        { path: '/albums/*', element: galleryPage },
        { path: '/files/*', element: filesPage },
        // { path: '/wormhole', element: wormholePage },
        { path: '/login', element: loginPage },
        { path: '/signup', element: signupPage },
        { path: '/setup', element: setupPage },
        { path: '/backup', element: backupPage },
        { path: '/settings/*', element: settingsPage },
        { path: '*', element: fourOhFourPage },
    ])

    return Gal || <Fourohfour />
}

function App() {
    const { theme, isOSControlled, changeTheme } = useWlTheme()
    useEffect(() => {
        document.documentElement.style.overflow = 'hidden'
        document.body.className = 'body'
    })

    const toggleThemeCb = useCallback(() => {
        if (isOSControlled) {
            return
        }
        changeTheme(
            theme === ThemeStateEnum.DARK
                ? ThemeStateEnum.LIGHT
                : ThemeStateEnum.DARK
        )
    }, [theme, isOSControlled, changeTheme])

    useKeyDown('t', toggleThemeCb)

    const queryClient = new QueryClient()

    return (
        <div className="flex h-screen w-screen">
            <Messages />
            <QueryClientProvider client={queryClient}>
                <Router>
                    <WeblensRoutes />
                </Router>
            </QueryClientProvider>
        </div>
    )
}

export default App
