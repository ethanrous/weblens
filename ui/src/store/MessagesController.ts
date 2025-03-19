import { StateCreator, create } from 'zustand'

export type Message = {
    text: string // Message to display
    duration: number // Length of time in ms to show the message

    id?: number // ID of the message
    title?: string // Title of the message
    queueTime?: number // Time the message was added to the queue
    severity?: 'info' | 'warning' | 'error' | 'debug' | 'success' // Severity of the message

    expired?: boolean
}

export interface MessagesStateT {
    messages: Map<number, Message>
    idCounter: number
    addMessage: (message: Message) => void
}

const MessagesControl: StateCreator<MessagesStateT, [], []> = (set) => ({
    messages: new Map(),
    idCounter: 0,

    addMessage: (message: Message) => {
        set((state) => {
            if (!message.severity) {
                message.severity = 'info'
            }

            message.queueTime = Date.now()
            const messageId = state.idCounter
            message.id = state.idCounter

            setTimeout(() => {
                set((laterState) => {
                    // laterState.messages.delete(messageId)
                    const message = laterState.messages.get(messageId)
                    message.expired = true
                    laterState.messages.set(messageId, message)
                    return {
                        ...laterState,
                        messages: new Map(laterState.messages),
                    }
                })
            }, message.duration)

            state.messages.set(messageId, message)

            return {
                ...state,
                idCounter: (state.idCounter + 1) % 1000,
                messages: new Map(state.messages),
            }
        })
    },
})

export const useMessagesController = create<MessagesStateT>()(MessagesControl)
