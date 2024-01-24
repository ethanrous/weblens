import React, { Suspense } from 'react'
import { BrowserRouter as Router, useRoutes } from "react-router-dom"
import { AppShell, MantineProvider, Text } from '@mantine/core'
import { Notifications } from '@mantine/notifications'

import WeblensLoader from "./components/Loading"
import Admin from "./Pages/Admin Settings/Admin"
import useR from "./components/UserInfo"
import { userContext } from "./Context"
import Login from "./Pages/Login/Login"

import "@mantine/notifications/styles.css"
import "@mantine/core/styles.css"
import { FlexColumnBox, FlexRowBox } from './Pages/FileBrowser/FilebrowserStyles'

const Gallery = React.lazy(() => import("./Pages/Gallery/Gallery"))
const FileBrowser = React.lazy(() => import("./Pages/FileBrowser/FileBrowser"))

const WeblensRoutes = () => {
  const { authHeader, userInfo, setCookie, clear } = useR()

  const galleryPage = (
    <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
      <Gallery />
    </Suspense>
  )

  const filesPage = (
    <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
      <FileBrowser />
    </Suspense>
  )

  const loginPage = (
    <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
      <Login />
    </Suspense>
  )

  const adminPage = (
    <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
      <Admin />
    </Suspense>
  )

  const Gal = useRoutes(
    [
      ...["/", "/timeline", "/albums/*"].map(path => ({ path: path, element: galleryPage })),
      { path: "/files/*", element: filesPage },
      { path: "/login", element: loginPage },
      { path: "/admin", element: adminPage },
    ]

  )
  return (
    <userContext.Provider value={{ authHeader, userInfo, setCookie, clear }}>
      {Gal}
    </userContext.Provider>
  )
}

function App() {
  // document.body.style.backgroundColor = theme.colorSchemes.dark.palette.neutral.solidDisabledBg
  document.documentElement.style.overflow = "hidden"
  // document.body.style.backgroundColor = "#fff"
  return (
    <MantineProvider defaultColorScheme="dark">
      <Notifications position='top-right' top={90} />
      <Router>
        <WeblensRoutes />
      </Router>
    </MantineProvider>
  )
}

export default App
