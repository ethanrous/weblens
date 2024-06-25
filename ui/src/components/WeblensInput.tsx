import { memo, ReactNode, useEffect, useRef, useState } from 'react'
import { useIsFocused, useKeyDown } from './hooks'
import WeblensButton from './WeblensButton'

import './weblensInput.scss'

const WeblensInput = memo(
    ({
        onComplete,
        value,
        valueCallback,
        Icon,
        buttonIcon,
        height,
        placeholder,
        openInput,
        closeInput,
        autoFocus = false,
        stealFocus = false,
        minimize = false,
    }: {
        onComplete: (v: string) => void
        value?: string
        valueCallback?: (v: string) => void
        Icon?: (p: any) => ReactNode
        buttonIcon?: (p: any) => ReactNode
        height?: number
        placeholder?: string
        openInput?: () => void
        closeInput?: () => void
        autoFocus?: boolean
        stealFocus?: boolean
        minimize?: boolean
    }) => {
        const searchRef = useRef<HTMLInputElement>()
        const isFocused = useIsFocused(searchRef.current)

        const [internalValue, setInternalValue] = useState(
            value !== undefined ? value : ''
        )
        useEffect(() => {
            setInternalValue(value !== undefined ? value : '')
        }, [value])

        useEffect(() => {
            if (isFocused && openInput) {
                openInput()
            } else if (!isFocused && closeInput) {
                closeInput()
            }
        }, [isFocused])

        useKeyDown('Enter', () => {
            if (onComplete) {
                onComplete(internalValue)
                if (closeInput) {
                    closeInput()
                }
            }
        })

        useKeyDown('Escape', (e) => {
            searchRef.current.blur()
        })

        useKeyDown(
            (e) => {
                return (
                    stealFocus &&
                    !e.metaKey &&
                    ((e.which >= 65 && e.which <= 90) || e.key === 'Backspace')
                )
            },
            (e) => {
                e.stopPropagation()
                searchRef.current.focus()
            }
        )

        return (
            <div
                className="weblens-input-wrapper"
                style={{ height: height, minWidth: height }}
                data-value={internalValue}
                data-minimize={minimize}
                onClick={() => {
                    searchRef.current.focus()
                }}
                onBlur={(e) => {
                    if (
                        closeInput &&
                        !e.currentTarget.contains(e.relatedTarget)
                    ) {
                        closeInput()
                    }
                }}
            >
                {Icon && <Icon className="w-max h-max" />}
                <input
                    ref={searchRef}
                    autoFocus={autoFocus}
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
                            Left={buttonIcon}
                            onClick={(e) => {
                                e.stopPropagation()
                                onComplete(internalValue)
                                if (closeInput) {
                                    closeInput()
                                }
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
        } else if (prev.closeInput !== next.closeInput) {
            return false
        }
        return true
    }
)

export default WeblensInput
