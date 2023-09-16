import Gallery from "./Pages/Gallery/Gallery"
import FileBrowser from "./Pages/FileBrowser/FileBrowser";
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
              <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
                <Gallery />
              </SnackbarProvider>
            </div>
          }
        />
        <Route
          path="/files/*"
          element={
            <div className="container">
              <SnackbarProvider maxSnack={10} autoHideDuration={5000}>
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
