import Login from "./Pages/Login/Login"
import Test from "./components/Test"
import GetWebsocket from './api/Websocket'

import React, { Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { useSnackbar, SnackbarProvider } from 'notistack'
import { LinearProgress } from "@mui/material";

const Gallery = React.lazy(() => import("./Pages/Gallery/Gallery"));
const FileBrowser = React.lazy(() => import("./Pages/FileBrowser/FileBrowser"));

function App() {
  const { enqueueSnackbar } = useSnackbar()
  const { wsSend, lastMessage, readyState } = GetWebsocket(enqueueSnackbar)

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <Suspense fallback={<LinearProgress style={{ width: "100%", position: "absolute" }} />}>
              <div className="container">
                <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
                  <Gallery wsSend={wsSend} lastMessage={lastMessage} readyState={readyState} enqueueSnackbar={enqueueSnackbar} />
                </SnackbarProvider>
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
            <Suspense fallback={<LinearProgress style={{ width: "100%", position: "absolute" }} />}>
              <div className="container">
                <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
                  <FileBrowser wsSend={wsSend} lastMessage={lastMessage} readyState={readyState} enqueueSnackbar={enqueueSnackbar} />
                </SnackbarProvider>
              </div>
            </Suspense>
          }
        />
        <Route
          path="/test"
          element={
            <div className="container">
              <Test />
            </div>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
