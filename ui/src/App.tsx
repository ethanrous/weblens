import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React, { Suspense, useEffect } from 'react'
import { CookiesProvider } from 'react-cookie'
import {
    BrowserRouter as Router,
    useLocation,
    useNavigate,
    useRoutes,
} from 'react-router-dom'
import ErrorBoundary from './components/Error'

import useR, { useSessionStore } from './components/UserInfo'
import StartUp from './pages/Startup/StartupPage'
import { useMediaStore } from './types/media/MediaStateControl'
import '@mantine/core/styles.css'
import Backup from './pages/Backup/Backup'
import Logo from './components/Logo'
import axios from 'axios'
import MediaApi from './api/MediaApi'
import { MediaTypeInfo } from './api/swag'
import { ErrorHandler } from './types/Types'

const Gallery = React.lazy(() => import('./pages/Gallery/Gallery'))
const FileBrowser = React.lazy(() => import('./pages/FileBrowser/FileBrowser'))
// const Wormhole = React.lazy(() => import('./pages/FileBrowser/Wormhole'))
const Login = React.lazy(() => import('./pages/Login/Login'))
const Setup = React.lazy(() => import('./pages/Setup/Setup'))

axios.defaults.withCredentials = true

async function saveMediaTypeMap(setState: (typeMap: MediaTypeInfo) => void) {
    return MediaApi.getMediaTypes().then((r) => {
        setState(r.data)
        localStorage.setItem(
            'mediaTypeMap',
            JSON.stringify({ typeMap: r.data, time: Date.now() })
        )
    })
}

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
            if (state.returnTo !== undefined) {
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

        const typeMapStr = localStorage.getItem('mediaTypeMap')
        if (!typeMapStr) {
            saveMediaTypeMap(setTypeMap).catch(ErrorHandler)
        }

        try {
            const typeMap = JSON.parse(typeMapStr) as MediaTypeInfo

            // fetch type map every hour, just in case
            // if (
            //     !typeMap.typeMap.size ||
            //     !typeMap.time ||
            //     Date.now() - typeMap.time > 3_600_000
            // ) {
            //     saveMediaTypeMap(setTypeMap)
            // }
            setTypeMap(typeMap)
        } finally {
            saveMediaTypeMap(setTypeMap).catch(ErrorHandler)
        }
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
        <div className="flex flex-col items-center justify-center h-screen w-screen gap-6">
            <h2 className="text-red-500">Fatal Error</h2>
            <h3>{err}</h3>

            <Logo />
        </div>
    )
}

const PageSwitcher = () => {
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

    const Gal = useRoutes([
        { path: '/', element: galleryPage },
        { path: '/timeline', element: galleryPage },
        { path: '/albums/*', element: galleryPage },
        { path: '/files/*', element: filesPage },
        // { path: '/wormhole', element: wormholePage },
        { path: '/login', element: loginPage },
        { path: '/setup', element: setupPage },
        { path: '/backup', element: backupPage },
    ])

    return Gal
}

function App() {
    document.documentElement.style.overflow = 'hidden'
    document.body.className = 'body'

    return (
        <MantineProvider defaultColorScheme="dark">
            <CookiesProvider defaultSetOptions={{ path: '/' }}>
                <Router>
                    <WeblensRoutes />
                </Router>
            </CookiesProvider>
        </MantineProvider>
    )
}

export default App
