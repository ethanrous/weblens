import { Text } from '@mantine/core'
import { Component } from 'react'
import { WeblensButton } from './WeblensButton'
import { useNavigate } from 'react-router-dom'

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
        // Example "componentStack":
        //   in ComponentThatThrows (created by App)
        //   in ErrorBoundary (created by App)
        //   in div (created by App)
        //   in App
        console.error(error, info.componentStack)
    }

    render() {
        if (this.state.hasError) {
            // You can render any custom fallback UI
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
        <div className="flex flex-col h-screen w-screen items-center justify-center">
            <Text size={'20px'} fw={600}>
                Something went wrong
            </Text>
            <Text style={{ margin: 10 }}>The error has been recorded</Text>
            <WeblensButton
                label="Go Back"
                squareSize={40}
                width={210}
                onClick={() => {
                    clearError()
                    nav(-1)
                }}
            />
        </div>
    )
}

export default ErrorBoundary
