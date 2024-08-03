import { MantineProvider } from '@mantine/core'
import React, { Suspense, useEffect, useReducer } from 'react'
import {
    BrowserRouter as Router,
    useLocation,
    useNavigate,
    useRoutes,
} from 'react-router-dom'
import { fetchMediaTypes } from './api/ApiFetch'
import ErrorBoundary, { ErrorDisplay } from './components/Error'

import WeblensLoader from './components/Loading'
import useR, { useSessionStore } from './components/UserInfo'
import { MediaContext } from './Context'

import '@mantine/notifications/styles.css'
import '@mantine/core/styles.css'
import { mediaReducer, MediaStateT } from './Media/Media'
import StartUp from './Pages/Startup/StartupPage'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CookiesProvider } from 'react-cookie'

const Gallery = React.lazy(() => import('./Pages/Gallery/Gallery'))
const FileBrowser = React.lazy(() => import('./Pages/FileBrowser/FileBrowser'))
const Wormhole = React.lazy(() => import('./Pages/FileBrowser/Wormhole'))
const Login = React.lazy(() => import('./Pages/Login/Login'))
const Setup = React.lazy(() => import('./Pages/Setup/Setup'))

const setTypeMap = () => {
    fetchMediaTypes().then((r) => {
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

    const loc = useLocation()
    const nav = useNavigate()

    useEffect(() => {
        fetchServerInfo()
    }, [])

    useEffect(() => {
        if (!server) {
            return
        }

        console.log(user)
        if (loc.pathname !== '/setup' && server.info.role === 'init') {
            console.debug('Nav setup')
            nav('/setup')
        } else if (loc.pathname === '/setup' && server.info.role !== 'init') {
            console.debug('Nav timeline')
            nav('/timeline')
        } else if (
            server.info.role === 'backup' &&
            !loc.pathname.startsWith('/files') &&
            user?.isLoggedIn
        ) {
            console.debug('Nav files home')
            nav('/files/home')
        } else if (
            !user?.isLoggedIn &&
            loc.pathname !== '/login' &&
            server.info.role !== 'init' &&
            !loc.pathname.startsWith('/files') &&
            loc.pathname === '/files/home'
        ) {
            console.debug('Nav login')
            nav('/login')
        } else if (loc.pathname === '/login' && user?.isLoggedIn) {
            console.debug('Nav timeline')
            nav('/timeline')
        } else if (
            (loc.pathname === '/timeline' ||
                loc.pathname.startsWith('/album')) &&
            server.info.role === 'backup'
        ) {
            console.debug('Nav files home')
            nav('/files/home')
        } else if (loc.pathname === '/') {
            console.debug('Nav timeline')
            nav('/timeline')
        }
    }, [loc, server, user])

    const [mediaState, mediaDispatch] = useReducer(
        mediaReducer,
        null,
        () =>
            new MediaStateT(
                null,
                JSON.parse(localStorage.getItem('showRaws')) || false,
                JSON.parse(localStorage.getItem('showHidden')) || false
            )
    )

    useEffect(() => {
        if (!server || !server.started || server.info.role === 'init') {
            return
        }

        const typeMapStr = localStorage.getItem('mediaTypeMap')
        if (!typeMapStr) {
            setTypeMap()
        }

        try {
            const typeMap = JSON.parse(typeMapStr)

            // fetch type map every hour, just in case
            if (
                !typeMap.typeMap.size ||
                !typeMap.time ||
                Date.now() - typeMap.time > 3_600_000
            ) {
                setTypeMap()
            }
        } catch {
            setTypeMap()
        }
    }, [])

    if (!server) {
        return null
    }

    const queryClient = new QueryClient()

    return (
        <QueryClientProvider client={queryClient}>
            <ErrorBoundary fallback={ErrorDisplay}>
                <MediaContext.Provider
                    value={{
                        mediaState,
                        mediaDispatch,
                    }}
                >
                    {server.started && <PageSwitcher />}
                    {!server.started && <StartUp />}
                </MediaContext.Provider>
            </ErrorBoundary>
        </QueryClientProvider>
    )
}

function PageLoader() {
    return (
        <div style={{ position: 'absolute', right: 15, bottom: 10 }}>
            <WeblensLoader loading={['page']} progress={0} />
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

    const Gal = useRoutes([
        { path: '/', element: galleryPage },
        { path: '/timeline', element: galleryPage },
        { path: '/albums/*', element: galleryPage },
        { path: '/files/*', element: filesPage },
        { path: '/wormhole', element: wormholePage },
        { path: '/login', element: loginPage },
        { path: '/setup', element: setupPage },
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
