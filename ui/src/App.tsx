import Login from "./Pages/Login/Login"
import React, { Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from "react-router-dom"
import { SnackbarProvider } from 'notistack'

import WeblensLoader from "./components/Loading";
import { ThemeProvider, createTheme } from "@mui/material";
const Gallery = React.lazy(() => import("./Pages/Gallery/Gallery"));
const FileBrowser = React.lazy(() => import("./Pages/FileBrowser/FileBrowser"));

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: "#ffffff",
      contrastText: "#ffffff"
    },
    secondary: {
      main: "#101418",
      dark: "#000010",
      contrastText: "#ffffff"
    },
    background: {
      default: '#101418',
    },
  },
  spacing: (val) => val,
  shape: {
    borderRadius: 6,
  }
});

const WeblensRoutes = () => {
  return (
    <BrowserRouter>
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
          path="/files/*"
          element={
            <Suspense fallback={<WeblensLoader loading={true} progress={0} />}>
              <div className="container">
                <FileBrowser />
              </div>
            </Suspense>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />
        }
        />
      </Routes>
    </BrowserRouter>
  )
}

function App() {
  document.body.style.backgroundColor = theme.palette.secondary.main
  // document.body.style = 'background: rgb(25, 25, 25);';
  // this.renderer.setStyle()

  return (
    <ThemeProvider theme={theme}>
      <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
        <WeblensRoutes />
      </SnackbarProvider >
    </ThemeProvider>
  );
}

export default App;
