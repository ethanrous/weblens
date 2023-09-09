import HomePage from "./components/HomePage"
import FileBrowser from "./components/FileBrowser";
import Test from "./components/Test";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { SnackbarProvider } from 'notistack';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <div className="container">
              <SnackbarProvider maxSnack={3} autoHideDuration={5000}>
                <HomePage />
              </SnackbarProvider>
            </div>
          }
        />
        <Route
          path="/files/*"
          element={
            <div className="container">
              <SnackbarProvider maxSnack={3} autoHideDuration={5000}>
                <FileBrowser />
              </SnackbarProvider>
            </div>
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
