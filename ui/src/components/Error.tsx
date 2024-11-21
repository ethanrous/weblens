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
        <div className="flex flex-col h-screen w-screen items-center justify-center theme-background">
            <div className="flex flex-col w-max">
                <p className="text-xl font-semibold m-2">
                    Sorry, something went wrong
                </p>
                <div className="flex flex-row w-full">
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
            <div className="flex flex-col w-[20vw] mt-16">
                <div
                    className="flex flex-row items-center cursor-pointer"
                    onClick={() => setErrOpen((o) => !o)}
                >
                    <p className="flex w-full justify-end">Advanced</p>
                    {!errOpen && (
                        <IconChevronLeft
                            size={20}
                            className="text-[--wl-text-color]"
                        />
                    )}
                    {errOpen && (
                        <IconChevronDown
                            size={20}
                            className="text-[--wl-text-color]"
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
