import {
    Message,
    useMessagesController,
} from '@weblens/store/MessagesController'
import { CSSProperties } from 'react'

export default function Messages() {
    const messages = useMessagesController((state) => state.messages)
    return (
        <div className="absolute right-4 bottom-4 z-50 w-max">
            {messages
                .values()
                .toArray()
                .sort((a, b) => a.queueTime - b.queueTime)
                .map((m, i) => (
                    <SingleMessage key={i} message={m} />
                ))}
        </div>
    )
}

function SingleMessage({ message }: { message: Message }) {
    let messageColor = ''
    switch (message.severity) {
        case 'info':
            messageColor = '--color-theme-primary'
            break
        case 'success':
            messageColor = '--color-valid'
            break
        case 'warning':
            messageColor = '--color-theme-warning'
            break
        case 'error':
            messageColor = '--color-button-danger'
            break
        case 'debug':
            messageColor = '--color-background-secondary'
            break
    }

    return (
        <div
            className="animate-fade-in border-message bg-background-primary m-2 rounded-md border transition"
            style={
                {
                    opacity: message.expired ? 0 : 100,
                    '--color-message': `var(${messageColor})`,
                } as CSSProperties
            }
        >
            <div className="bg-message/50 p-2">
                <h4>{message.title}</h4>
                <span>{message.text}</span>
            </div>
        </div>
    )
}
