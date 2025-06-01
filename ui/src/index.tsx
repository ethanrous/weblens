import { Suspense, lazy } from 'react'
import ReactDOM from 'react-dom/client'
// import './components/theme.module.scss'
import '~/css/main.css'

import WeblensLoader from './components/Loading'

const App = lazy(() => import('./App'))

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
    <div className="bg-wl-background flex h-screen w-screen overflow-hidden">
        <Suspense fallback={<WeblensLoader className="m-auto" />}>
            <App />
        </Suspense>
    </div>
)
