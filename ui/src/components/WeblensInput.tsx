import { memo, useEffect, useState } from 'react'
import { useKeyDown } from './hooks'

const WeblensInput = memo(
    ({
        value,
        onComplete,
        valueCallback,
        closeInput,
    }: {
        value: string
        onComplete: (v: string) => void
        valueCallback?: (v: string) => void
        closeInput?: () => void
    }) => {
        const [inputValue, setInputValue] = useState(value)
        useKeyDown('Enter', () => onComplete(inputValue))

        useEffect(() => {
            if (valueCallback) {
                valueCallback(inputValue)
            }
        }, [inputValue])

        return (
            <input
                autoFocus
                value={inputValue}
                className="weblens-input"
                onChange={(event) => setInputValue(event.target.value)}
                onBlur={closeInput}
            />
        )
    },
    (prev, next) => {
        if (prev.value !== next.value) {
            return false
        } else if (prev.onComplete !== next.onComplete) {
            return false
        }
        return true
    }
)

export default WeblensInput
