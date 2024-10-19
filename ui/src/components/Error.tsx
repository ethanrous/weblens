import { Text } from '@mantine/core'
import { useWebsocketStore } from '@weblens/api/Websocket'
import WeblensButton from '@weblens/lib/WeblensButton'
import { Component } from 'react'
import { useNavigate } from 'react-router-dom'

class ErrorBoundary extends Component<{ children }, { hasError: boolean }> {
    constructor(props) {
        super(props)
        this.state = { hasError: props.hasError }
    }

    static getDerivedStateFromError() {
        // Update state so the next render will show the fallback UI.
        return { hasError: true }
    }

    clearError() {
        // Update state so the next render will show the fallback UI.
        this.setState({ hasError: false })
    }

    componentDidCatch(error) {
        const wsSend = useWebsocketStore.getState().wsSend
        if (wsSend != null) {
            wsSend(
                JSON.stringify({
                    action: 'show_web_error',
                    content: error.message,
                })
            )
        }
    }

    render() {
        if (this.state.hasError) {
            return (
                <ErrorDisplay
                    clearError={() => {
                        this.clearError()
                    }}
                />
            )
        }

        return this.props.children
    }
}

function ErrorDisplay({ clearError }) {
    const nav = useNavigate()
    return (
        <div className="flex flex-col h-screen w-screen items-center justify-center theme-background">
            <p className="text-xl font-semibold">Something went wrong</p>
            <p style={{ margin: 10 }}>The error has been recorded</p>
            <div className="flex flex-row w-max">
                <WeblensButton
                    label="Go Home"
                    centerContent
                    squareSize={40}
                    onClick={() => {
                        clearError()
                        nav('/')
                    }}
                />
                <WeblensButton
                    label="Refresh"
                    centerContent
                    squareSize={40}
                    onClick={() => {
                        clearError()
                        location.reload()
                    }}
                />
            </div>
        </div>
    )
}

export default ErrorBoundary
