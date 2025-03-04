import { MantineProvider } from '@mantine/core'
import '@mantine/core/styles.css'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import axios from 'axios'
import React, { Suspense, useCallback, useEffect } from 'react'
import {
    BrowserRouter as Router,
    useLocation,
    useNavigate,
    useRoutes,
} from 'react-router-dom'

import MediaApi from './api/MediaApi'
import ErrorBoundary from './components/Error'
import Logo from './components/Logo'
import Messages from './components/Messages'
import useR, { useSessionStore } from './components/UserInfo'
import { useKeyDown } from './components/hooks'
import Fourohfour from './pages/404/fourohfour'
import Backup from './pages/Backup/Backup'
import Signup from './pages/Signup/Signup'
import StartUp from './pages/Startup/StartupPage'
import { SettingsMenu } from './pages/UserSettings/Settings'
import { ThemeStateEnum, useWlTheme } from './store/ThemeControl'
import { ErrorHandler } from './types/Types'
import { useMediaStore } from './types/media/MediaStateControl'

const Gallery = React.lazy(() => import('./pages/Gallery/Gallery'))
const FileBrowser = React.lazy(() => import('./pages/FileBrowser/FileBrowser'))
const Login = React.lazy(() => import('./pages/Login/Login'))
const Setup = React.lazy(() => import('./pages/Setup/Setup'))

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
    }, [])

    useEffect(() => {
        if (nav) {
            setNav(nav)
        }
    }, [nav])

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
            user?.isLoggedIn
        ) {
            console.debug('Nav backup page')
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
    }, [loc, server, user])

    useEffect(() => {
        if (!server || !server.started || server.role === 'init') {
            return
        }
        MediaApi.getMediaTypes()
            .then((r) => {
                setTypeMap(r.data)
            })
            .catch(ErrorHandler)
    }, [server])

    useEffect(() => {
        const theme = localStorage.getItem('theme')
        if (theme === 'dark') {
            document.documentElement.classList.toggle('dark')
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
        return <PageLoader />
    }

    const queryClient = new QueryClient()

    return (
        <QueryClientProvider client={queryClient}>
            <ErrorBoundary>
                {server.started && <PageSwitcher />}
                {!server.started && <StartUp />}
            </ErrorBoundary>
        </QueryClientProvider>
    )
}

function PageLoader() {
    return (
        <div
            style={{
                position: 'absolute',
                right: 15,
                bottom: 10,
                opacity: 0.5,
            }}
        >
            <Logo />
        </div>
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
        <Suspense fallback={<PageLoader />}>
            <Gallery />
        </Suspense>
    )

    const filesPage = (user.isLoggedIn ||
        loc.pathname.startsWith('/files/share')) && (
        <Suspense fallback={<PageLoader />}>
            <FileBrowser />
        </Suspense>
    )

    // const wormholePage = (
    //     <Suspense fallback={<PageLoader />}>
    //         <Wormhole />
    //     </Suspense>
    // )

    const loginPage = (
        <Suspense fallback={<PageLoader />}>
            <Login />
        </Suspense>
    )

    const signupPage = (
        <Suspense fallback={<PageLoader />}>
            <Signup />
        </Suspense>
    )

    const setupPage = (
        <Suspense fallback={<PageLoader />}>
            <Setup />
        </Suspense>
    )

    const backupPage = (
        <Suspense fallback={<PageLoader />}>
            <Backup />
        </Suspense>
    )

    const settingsPage = (
        <Suspense fallback={<PageLoader />}>
            <SettingsMenu />
        </Suspense>
    )

    const fourOhFourPage = (
        <Suspense fallback={<PageLoader />}>
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
    document.documentElement.style.overflow = 'hidden'
    document.body.className = 'body'
    const { theme, isOSControlled, changeTheme } = useWlTheme()

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

    return (
        <MantineProvider defaultColorScheme="dark">
            <Messages />
            <Router>
                <WeblensRoutes />
            </Router>
        </MantineProvider>
    )
}

export default App
