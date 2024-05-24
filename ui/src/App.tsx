import React, { Suspense, useEffect } from "react";
import { BrowserRouter as Router, useRoutes } from "react-router-dom";
import { Box, MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";

import WeblensLoader from "./components/Loading";
import useR from "./components/UserInfo";
import { UserContext } from "./Context";

import "@mantine/notifications/styles.css";
import "@mantine/core/styles.css";
import { fetchMediaTypes } from "./api/ApiFetch";
import ErrorBoundary, { ErrorDisplay } from "./components/Error";

const Gallery = React.lazy(() => import("./Pages/Gallery/Gallery"));
const FileBrowser = React.lazy(() => import("./Pages/FileBrowser/FileBrowser"));
const Wormhole = React.lazy(() => import("./Pages/FileBrowser/Wormhole"));
const Login = React.lazy(() => import("./Pages/Login/Login"));
const Setup = React.lazy(() => import("./Pages/Setup/Setup"));

const setTypeMap = () => {
    fetchMediaTypes().then((r) => {
        localStorage.setItem(
            "mediaTypeMap",
            JSON.stringify({ typeMap: r, time: Date.now() })
        );
    });
};

const WeblensRoutes = () => {
    const { authHeader, usr, setCookie, clear, serverInfo } = useR();

    useEffect(() => {
        const typeMapStr = localStorage.getItem("mediaTypeMap");
        if (!typeMapStr) {
            setTypeMap();
        }

        try {
            const typeMap = JSON.parse(typeMapStr);

            // fetch type map every hour, just in case
            if (!typeMap.time || Date.now() - typeMap.time > 3_600_000) {
                setTypeMap();
            }
        } catch {
            setTypeMap();
        }
    }, []);
    // const [typeMap, setTypeMap] = useState(null);
    // useEffect(() => {
    //     const mediaTypes = new Map<string, mediaType>();
    //     fetchMediaTypes().then((mt) => {
    //         const mimes: string[] = Array.from(Object.keys(mt));
    //         for (const mime of mimes) {
    //             mediaTypes.set(mime, mt[mime]);
    //         }
    //         setTypeMap(mediaTypes);
    //     });
    // }, []);

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
        ...["/", "/timeline", "/albums/*"].map((path) => ({
            path: path,
            element: galleryPage,
        })),
        { path: "/files/*", element: filesPage },
        { path: "/share/*", element: filesPage },
        { path: "/wormhole/*", element: wormholePage },
        { path: "/login", element: loginPage },
        { path: "/setup", element: setupPage },
    ]);

    return (
        <ErrorBoundary fallback={ErrorDisplay}>
            <UserContext.Provider
                value={{
                    authHeader,
                    usr,
                    setCookie,
                    clear,
                    serverInfo,
                }}
            >
                {/* <MediaTypeContext.Provider value={typeMap}> */}
                {Gal}
                {/* </MediaTypeContext.Provider> */}
            </UserContext.Provider>
        </ErrorBoundary>
    );
};

const PageLoader = () => {
    return (
        <Box style={{ position: "absolute", right: 15, bottom: 10 }}>
            <WeblensLoader loading={["page"]} progress={0} />
        </Box>
    );
};

function App() {
    document.documentElement.style.overflow = "hidden";
    document.body.className = "body";
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
