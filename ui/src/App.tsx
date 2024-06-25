import { MantineProvider } from '@mantine/core'
import { Notifications } from '@mantine/notifications'
import React, { memo, Suspense, useEffect, useReducer } from 'react'
import { BrowserRouter as Router, useRoutes } from 'react-router-dom'
import { fetchMediaTypes } from './api/ApiFetch'
import ErrorBoundary, { ErrorDisplay } from './components/Error'

import WeblensLoader from './components/Loading'
import useR from './components/UserInfo'
import { MediaContext, UserContext } from './Context'

import '@mantine/notifications/styles.css'
import '@mantine/core/styles.css'
import { MediaAction, mediaReducer, MediaStateT } from './Media/Media'

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
    const { authHeader, usr, setCookie, clear, serverInfo } = useR()

    useEffect(() => {
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

    const [mediaState, mediaDispatch] = useReducer<
        (state: MediaStateT, action: MediaAction) => MediaStateT
    >(mediaReducer, new MediaStateT())

    return (
        <ErrorBoundary fallback={ErrorDisplay}>
            <MediaContext.Provider
                value={{
                    mediaState,
                    mediaDispatch,
                }}
            >
                <UserContext.Provider
                    value={{
                        authHeader,
                        usr,
                        setCookie,
                        clear,
                        serverInfo,
                    }}
                >
                    <PageSwitcher />
                </UserContext.Provider>
            </MediaContext.Provider>
        </ErrorBoundary>
    )
}

function PageLoader() {
    return (
        <div style={{ position: 'absolute', right: 15, bottom: 10 }}>
            <WeblensLoader loading={['page']} progress={0} />
        </div>
    )
}

const PageSwitcher = memo(
    () => {
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
            ...['/', '/timeline', '/albums/*'].map((path) => ({
                path: path,
                element: galleryPage,
            })),
            { path: '/files/*', element: filesPage },
            // { path: "/share/*", element: filesPage },
            { path: '/wormhole/*', element: wormholePage },
            { path: '/login', element: loginPage },
            { path: '/setup', element: setupPage },
        ])

        return Gal
    },
    (prev, next) => {
        return true
    }
)

function App() {
    document.documentElement.style.overflow = 'hidden'
    document.body.className = 'body'
    return (
        <MantineProvider defaultColorScheme="dark">
            <Notifications position="top-right" top={90} />
            <Router>
                <WeblensRoutes />
            </Router>
        </MantineProvider>
    )
}

export default App
