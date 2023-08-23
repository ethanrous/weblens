import HomePage from "./components/HomePage"
import Upload from "./components/Upload";
import { BrowserRouter, Routes, Route } from "react-router-dom";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <div className="container">
              <HomePage />
            </div>
          }
        />
        <Route
          path="/upload"
          element={
            <div className="container">
              <Upload />
            </div>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
