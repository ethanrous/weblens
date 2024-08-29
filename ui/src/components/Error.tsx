import { Text } from '@mantine/core'
import { Component } from 'react'
import WeblensButton from './WeblensButton'
import { useNavigate } from 'react-router-dom'
import { useWebsocketStore } from '../api/Websocket'

class ErrorBoundary extends Component<
    { children; fallback: (p) => JSX.Element },
    { hasError: boolean }
> {
    constructor(props) {
        super(props)
        this.state = { hasError: props.hasError }
    }

    static getDerivedStateFromError(error) {
        // Update state so the next render will show the fallback UI.
        return { hasError: true }
    }

    clearError() {
        // Update state so the next render will show the fallback UI.
        this.setState({ hasError: false })
    }

    componentDidCatch(error, info) {
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

export function ErrorDisplay({ clearError }) {
    const nav = useNavigate()
    return (
        <div className="flex flex-col h-screen w-screen items-center justify-center bg-bottom-grey">
            <Text size={'20px'} fw={600}>
                Something went wrong
            </Text>
            <Text style={{ margin: 10 }}>The error has been recorded</Text>
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
