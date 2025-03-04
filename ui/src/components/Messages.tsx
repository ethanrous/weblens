import {
    Message,
    useMessagesController,
} from '@weblens/store/MessagesController'
import { CSSProperties } from 'react'

export default function Messages() {
    const messages = useMessagesController((state) => state.messages)
    return (
        <div className="absolute bottom-4 right-4 z-50 w-max">
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
            messageColor = '--wl-theme-color-primary'
            break
        case 'success':
            messageColor = '--wl-color-valid'
            break
        case 'warning':
            messageColor = '--wl-theme-color-warning'
            break
        case 'error':
            messageColor = '--wl-button-color-danger'
            break
        case 'debug':
            messageColor = '--wl-background-color-secondary'
            break
    }

    return (
        <div
            className="m-2 animate-fade rounded-md border border-wl-message-color transition bg-wl-background-color-primary"
            style={
                {
                    opacity: message.expired ? 0 : 100,
                    '--wl-message-color': `var(${messageColor})`,
                } as CSSProperties
            }
        >
			<div className='bg-wl-message-color/20 p-2'>
				<h4>{message.title}</h4>
				<span>{message.text}</span>
			</div>
        </div>
    )
}
