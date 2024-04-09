import React, { Suspense } from 'react';
import { BrowserRouter as Router, useRoutes } from 'react-router-dom';
import { Box, MantineProvider } from '@mantine/core';
import { Notifications } from '@mantine/notifications';

import WeblensLoader from './components/Loading';
import useR from './components/UserInfo';
import { userContext } from './Context';

import '@mantine/notifications/styles.css';
import '@mantine/core/styles.css';

const Gallery = React.lazy(() => import('./Pages/Gallery/Gallery'));
const FileBrowser = React.lazy(() => import('./Pages/FileBrowser/FileBrowser'));
const Wormhole = React.lazy(() => import('./Pages/FileBrowser/Wormhole'));
const Login = React.lazy(() => import('./Pages/Login/Login'));
const Setup = React.lazy(() => import('./Pages/Setup/Setup'));

const WeblensRoutes = () => {
    const { authHeader, usr, setCookie, clear } = useR();

    const galleryPage = (
        <Suspense fallback={<PageLoader />}>
            <Gallery />
        </Suspense>
    );

    const filesPage = (
        <Suspense fallback={<PageLoader />}>
            <FileBrowser />
        </Suspense>
    );

    const wormholePage = (
        <Suspense fallback={<PageLoader />}>
            <Wormhole />
        </Suspense>
    );

    const loginPage = (
        <Suspense fallback={<PageLoader />}>
            <Login />
        </Suspense>
    );

    const setupPage = (
        <Suspense fallback={<PageLoader />}>
            <Setup />
        </Suspense>
    );

    const Gal = useRoutes([
        ...['/', '/timeline', '/albums/*'].map((path) => ({
            path: path,
            element: galleryPage,
        })),
        { path: '/files/*', element: filesPage },
        { path: '/share/*', element: filesPage },
        { path: '/wormhole/*', element: wormholePage },
        { path: '/login', element: loginPage },
        { path: '/setup', element: setupPage },
    ]);
    return (
        <userContext.Provider value={{ authHeader, usr, setCookie, clear }}>
            {Gal}
        </userContext.Provider>
    );
};

const PageLoader = () => {
    return (
        <Box style={{ position: 'absolute', right: 15, bottom: 10 }}>
            <WeblensLoader loading={['page']} progress={0} />
        </Box>
    );
};

function App() {
    // document.body.style.backgroundColor = theme.colorSchemes.dark.palette.neutral.solidDisabledBg
    document.documentElement.style.overflow = 'hidden';
    document.body.style.backgroundColor = '#111418';
    return (
        <MantineProvider defaultColorScheme="dark">
            <Notifications position="top-right" top={90} />
            <Router>
                <WeblensRoutes />
            </Router>
        </MantineProvider>
    );
}

export default App;
