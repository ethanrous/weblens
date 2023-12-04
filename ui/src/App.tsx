import React, { Suspense } from 'react'
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { SnackbarProvider } from 'notistack'
import Fourohfour from "./Pages/404/fourohfour"
import { Box, Button, CssVarsProvider, extendTheme } from "@mui/joy"
import { MantineProvider } from '@mantine/core'
import { Notifications, notifications } from '@mantine/notifications'

import WeblensLoader from "./components/Loading"
import Admin from "./Pages/Admin Settings/Admin"
import useR from "./components/UserInfo"
import { userContext } from "./Context"
import SignUp from "./Pages/Login/SignUp"
import Login from "./Pages/Login/Login"

import "@mantine/notifications/styles.css"
import "@mantine/core/styles.css"

const Gallery = React.lazy(() => import("./Pages/Gallery/Gallery"))
const FileBrowser = React.lazy(() => import("./Pages/FileBrowser/FileBrowser"))

const palette = {
  text: {
    primary: '#fff',
    plain: '#fff',
    icon: '#fff'
  },
  primary: {
    outlinedColor: 'rgb(51, 51, 153)',

    solidBg: 'rgb(50 50 55)',
    solidActiveBg: 'rgb(30 30 90)',
    solidDisabledBg: 'rgb(30 30 35)',

    softBg: 'rgba(20 00 75 / 0.5)',

    plainColor: '#442299',

  },
  neutral: {
    solidBg: 'rgb(25 15 55)',
    // solidColor: '#fff',
    solidColor: '#fff',
    plainColor: '#aa88ff',
    solidDisabledBg: '#111418',
    solidActiveBg: '#111418',
    // mainChannel: '255 255 255',
    mainChannel: '255 255 255',
    // softBg: 'rgba(25 25 45 / 0.80)',
    softBg: 'rgb(47 40 100)',
    outlinedColor: '#fff',

  },
  success: {
    solidBg: '#198754',
    solidBorder: '#198754',
    solidHoverBg: '#157347',
    solidHoverBorder: '#146c43',
    solidActiveBg: '#146c43',
    solidActiveBorder: '#13653f',
    solidDisabledBg: '#198754',
    solidDisabledBorder: '#198754',
  },
  danger: {

  },
  warning: {

  },
  info: {

  },
  background: {
    body: '#050522',
    surface: '#074aa155'
  }
};


const theme = extendTheme({
  colorSchemes: {
    light: { palette },
    dark: { palette },
  },
  fontFamily: {
    // display: "Roboto",
    // body: "Roboto, sans-serif;",
  },
  components: {
    JoyTypography: {
      defaultProps: {

      }
    },
    JoyInput: {
      defaultProps: {

      }
    },
    JoySheet: {
      styleOverrides: {
        root: ({ ownerState, theme }) => ({
          ...(ownerState.color === 'primary' && {
            backgroundColor: theme.vars.palette.background.surface,
          }),
        }),
      }
    },
    JoyButton: {
      styleOverrides: {
        root: ({ theme, ownerState }) => ({
          '&:focus': theme.focus.default,
          fontWeight: 600,
          ...(ownerState.size === 'md' && {
            borderRadius: '0.375rem',
            paddingInline: '1rem',
          }),
        }),
      },
    },
  }
});

const WeblensRoutes = () => {
  const { authHeader, userInfo, setCookie, removeCookie } = useR()

  return (
    <userContext.Provider value={{ authHeader, userInfo, setCookie, removeCookie }}>
      <Routes>
        <Route
          path="/"
          element={
            <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
              <div className="container">
                <Gallery />
              </div>
            </Suspense>
          }
        />
        <Route
          path="/login"
          element={
            <div className="container">
              <Login />
            </div>
          }
        />
        <Route
          path="/signup"
          element={
            <div className="container">
              <SignUp />
            </div>
          }
        />
        <Route
          path="/files/*"
          element={
            <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
              {/* <div className="container"> */}
              <FileBrowser />
              {/* </div> */}
            </Suspense>
          }
        />
        <Route
          path="/admin"
          element={
            <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
              <div className="container">
                <Admin />
              </div>
            </Suspense>
          }
        />
        <Route path="*" element={
          <Fourohfour />
        }
        />
      </Routes>
    </userContext.Provider>
  )
}

function App() {
  document.body.style.backgroundColor = theme.colorSchemes.dark.palette.neutral.solidDisabledBg
  document.documentElement.style.overflow = "hidden"
  // document.body.style.backgroundColor = "#fff"

  return (
    <MantineProvider defaultColorScheme="dark">
      <Notifications />
      <BrowserRouter>
        <CssVarsProvider
          defaultMode="system"
          theme={theme}
        >
          <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
            <WeblensRoutes />
          </SnackbarProvider >
        </CssVarsProvider>
      </BrowserRouter>
    </MantineProvider>
  );
}

export default App;
