import React, { Suspense } from 'react';
import { BrowserRouter as Router, useRoutes } from 'react-router-dom';
import { MantineProvider } from '@mantine/core';
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
const Admin = React.lazy(() => import('./Pages/Admin Settings/Admin'));

const WeblensRoutes = () => {
    const { authHeader, usr, setCookie, clear } = useR();

    const galleryPage = (
        <Suspense fallback={<WeblensLoader loading={['page']} progress={0} />}>
            <Gallery />
        </Suspense>
    );

    const filesPage = (
        <Suspense fallback={<WeblensLoader loading={['page']} progress={0} />}>
            <FileBrowser />
        </Suspense>
    );

    const wormholePage = (
        <Suspense fallback={<WeblensLoader loading={['page']} progress={0} />}>
            <Wormhole />
        </Suspense>
    );

    const loginPage = (
        <Suspense fallback={<WeblensLoader loading={['page']} progress={0} />}>
            <Login />
        </Suspense>
    );

    const adminPage = (
        <Suspense fallback={<WeblensLoader loading={['page']} progress={0} />}>
            <Admin />
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
        { path: '/admin', element: adminPage },
    ]);
    return (
        <userContext.Provider value={{ authHeader, usr, setCookie, clear }}>
            {Gal}
        </userContext.Provider>
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
