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

import WeblensLoader from './components/Loading'
import useR, { useSessionStore } from './components/UserInfo'
import StartUp from './pages/Startup/StartupPage'
import { fetchMediaTypes } from './types/media/MediaQuery'
import { useMediaStore } from './types/media/MediaStateControl'
import '@mantine/core/styles.css'
import Backup from './pages/Backup/Backup'

const Gallery = React.lazy(() => import('./pages/Gallery/Gallery'))
const FileBrowser = React.lazy(() => import('./pages/FileBrowser/FileBrowser'))
const Wormhole = React.lazy(() => import('./pages/FileBrowser/Wormhole'))
const Login = React.lazy(() => import('./pages/Login/Login'))
const Setup = React.lazy(() => import('./pages/Setup/Setup'))

const saveMediaTypeMap = (setState) => {
    fetchMediaTypes().then((r) => {
        setState(r)
        localStorage.setItem(
            'mediaTypeMap',
            JSON.stringify({ typeMap: r, time: Date.now() })
        )
    })
}

const WeblensRoutes = () => {
    useR()

    const fetchServerInfo = useSessionStore((state) => state.fetchServerInfo)
    const server = useSessionStore((state) => state.server)
    const user = useSessionStore((state) => state.user)

    const setNav = useSessionStore((state) => state.setNav)

    const loc = useLocation()
    const nav = useNavigate()

    const setTypeMap = useMediaStore((state) => state.setTypeMap)

    useEffect(() => {
        fetchServerInfo()
    }, [])

    useEffect(() => {
        if (nav) {
            setNav(nav)
        }
    }, [nav])

    useEffect(() => {
        if (!server) {
            return
        }
        // if (loc.pathname !== '/login' && user?.isLoggedIn === false) {
        //     console.debug('Nav login')
        //     nav('/login')
        // }
        else if (loc.pathname !== '/setup' && server.info.role === 'init') {
            console.debug('Nav setup')
            nav('/setup')
        } else if (loc.pathname === '/setup' && server.info.role === 'core') {
            console.debug('Nav files home')
            nav('/files/home')
        } else if (
            server.info.role === 'backup' &&
            loc.pathname !== '/backup' &&
            user?.isLoggedIn
        ) {
            console.debug('Nav backup page')
            nav('/backup')
        } else if (loc.pathname === '/login' && user?.isLoggedIn) {
            if (loc.state?.returnTo) {
                console.debug('Nav return to')
                nav(loc.state.returnTo)
            } else {
                console.debug('Nav files home')
                nav('/files/home')
            }
        } else if (
            (loc.pathname === '/timeline' ||
                loc.pathname.startsWith('/album')) &&
            server.info.role === 'backup'
        ) {
            console.debug('Nav files home')
            nav('/files/home')
        } else if (loc.pathname === '/' && server.info.role === 'core') {
            console.debug('Nav timeline')
            nav('/timeline')
        }
    }, [loc, server, user])

    useEffect(() => {
        if (!server || !server.started || server.info.role === 'init') {
            return
        }

        const typeMapStr = localStorage.getItem('mediaTypeMap')
        if (!typeMapStr) {
            saveMediaTypeMap(setTypeMap)
        }

        try {
            const typeMap = JSON.parse(typeMapStr)

            // fetch type map every hour, just in case
            if (
                !typeMap.typeMap.size ||
                !typeMap.time ||
                Date.now() - typeMap.time > 3_600_000
            ) {
                saveMediaTypeMap(setTypeMap)
            }
            setTypeMap(typeMap.typeMap)
        } catch {
            saveMediaTypeMap(setTypeMap)
        }
    }, [server])

    useEffect(() => {
        const theme = localStorage.getItem('theme')
        if (theme === 'dark') {
            document.documentElement.classList.toggle('dark')
        }
    }, [])

    if (!server || !user) {
        return null
    }

    const queryClient = new QueryClient()

    return (
        <QueryClientProvider client={queryClient}>
            <ErrorBoundary>
                {server.started && user && <PageSwitcher />}
                {!server.started && <StartUp />}
            </ErrorBoundary>
        </QueryClientProvider>
    )
}

function PageLoader() {
    return (
        <div style={{ position: 'absolute', right: 15, bottom: 10 }}>
            <WeblensLoader loading={['page']} />
        </div>
    )
}

const PageSwitcher = () => {
    const galleryPage = (
        <Suspense fallback={<PageLoader />}>
            <Gallery />
        </Suspense>
    )

    const filesPage = (
        <Suspense fallback={<PageLoader />}>
            <FileBrowser />
        </Suspense>
    )

    const wormholePage = (
        <Suspense fallback={<PageLoader />}>
            <Wormhole />
        </Suspense>
    )

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
        { path: '/wormhole', element: wormholePage },
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
