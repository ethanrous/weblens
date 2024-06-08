import { memo, ReactNode, useState } from 'react'
import { useKeyDown } from './hooks'
import { WeblensButton } from './WeblensButton'

const WeblensInput = memo(
    ({
        onComplete,
        value,
        valueCallback,
        icon,
        buttonIcon,
        height,
        placeholder,
        closeInput,
    }: {
        onComplete: (v: string) => void
        value?: string
        valueCallback?: (v: string) => void
        icon?
        buttonIcon?: (p: any) => ReactNode
        height?: number
        placeholder?: string
        closeInput?: () => void
    }) => {
        const [internalValue, setInternalValue] = useState(value ? value : '')
        useKeyDown('Enter', () => {
            if (onComplete) {
                onComplete(internalValue)
            }
        })

        return (
            <div
                className="weblens-input-wrapper"
                style={{ height: height }}
                onBlur={(e) => {
                    if (
                        closeInput &&
                        !e.currentTarget.contains(e.relatedTarget)
                    ) {
                        closeInput()
                    }
                }}
            >
                {icon}
                <input
                    autoFocus
                    className="weblens-input"
                    value={internalValue}
                    placeholder={placeholder}
                    onChange={(event) => {
                        if (valueCallback) {
                            valueCallback(event.target.value)
                        }
                        setInternalValue(event.target.value)
                    }}
                    onClick={(e) => e.stopPropagation()}
                />
                {buttonIcon && (
                    <div className="flex w-max justify-end" tabIndex={0}>
                        <WeblensButton
                            centerContent
                            squareSize={height ? height * 0.75 : 40}
                            width={height ? height * 0.75 : 40}
                            Left={buttonIcon}
                            onClick={(e) => {
                                e.stopPropagation()
                                console.log(internalValue)
                                onComplete(internalValue)
                            }}
                        />
                    </div>
                )}
            </div>
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
