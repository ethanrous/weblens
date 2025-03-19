import { IconChevronDown, IconChevronLeft } from '@tabler/icons-react'
import { WsMsgEvent, useWebsocketStore } from '@weblens/api/Websocket'
import WeblensButton from '@weblens/lib/WeblensButton'
import { Component, ReactNode, useState } from 'react'
import { useNavigate } from 'react-router-dom'

type ErrorBoundaryProps = {
    error: string
}

class ErrorBoundary extends Component<
    { children: ReactNode },
    ErrorBoundaryProps
> {
    constructor(props: { error: string; children: ReactNode }) {
        super(props)
        this.state = { error: props.error }
    }

    static getDerivedStateFromError() {
        // Update state so the next render will show the fallback UI.
        return { hasError: true }
    }

    clearError() {
        // Update state so the next render will show the fallback UI.
        this.setState({ error: null })
    }

    componentDidCatch(error: Error) {
        const wsSend = useWebsocketStore.getState().wsSend
        this.setState({ error: error.message })
        if (wsSend != null && error.message != null) {
            wsSend(WsMsgEvent.ErrorEvent, {
                action: 'show_web_error',
                content: error.message,
            })
        }
    }

    render() {
        if (this.state.error) {
            return (
                <ErrorDisplay
                    err={this.state.error}
                    clearError={() => {
                        this.clearError()
                    }}
                />
            )
        }

        return this.props.children
    }
}

function ErrorDisplay({
    err,
    clearError,
}: {
    err: string
    clearError: () => void
}) {
    const nav = useNavigate()
    const [errOpen, setErrOpen] = useState(false)
    return (
        <div className="theme-background flex h-screen w-screen flex-col items-center justify-center">
            <div className="flex w-max flex-col">
                <p className="m-2 text-xl font-semibold">
                    Sorry, something went wrong
                </p>
                <div className="flex w-full flex-row gap-2">
                    <WeblensButton
                        label="Go Home"
                        centerContent
                        fillWidth
                        squareSize={40}
                        onClick={() => {
                            clearError()
                            nav('/')
                        }}
                    />
                    <WeblensButton
                        label="Refresh"
                        centerContent
                        fillWidth
                        squareSize={40}
                        onClick={() => {
                            clearError()
                            location.reload()
                        }}
                    />
                </div>
            </div>
            <div className="mt-16 flex w-[20vw] flex-col">
                <div
                    className="flex cursor-pointer flex-row items-center"
                    onClick={() => setErrOpen((o) => !o)}
                >
                    <p className="flex w-full justify-end">Advanced</p>
                    {!errOpen && (
                        <IconChevronLeft
                            size={20}
                            className="text-(--color-text)"
                        />
                    )}
                    {errOpen && (
                        <IconChevronDown
                            size={20}
                            className="text-(--color-text)"
                        />
                    )}
                </div>
                <p
                    className="text-red-500"
                    style={{ visibility: errOpen ? 'visible' : 'hidden' }}
                >
                    {err}
                </p>
            </div>
        </div>
    )
}

export default ErrorBoundary
