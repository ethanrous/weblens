import WeblensButton from '@weblens/lib/WeblensButton'
import { memo, ReactNode, useEffect, useState } from 'react'
import { useIsFocused, useResize } from '@weblens/components/hooks'

import '@weblens/lib/weblensInput.scss'

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
        ignoreKeys = [],
        password = false,
        minimize = false,
        subtle = false,
        fillWidth = true,
        failed = false,
    }: {
        onComplete?: (v: string) => Promise<void | Response>
        value?: string
        valueCallback?: (v: string) => void
        Icon?: (p) => ReactNode
        buttonIcon?: (p) => ReactNode
        height?: number
        placeholder?: string
        openInput?: () => void
        closeInput?: () => void
        autoFocus?: boolean
        stealFocus?: boolean
        ignoreKeys?: string[]
        password?: boolean
        minimize?: boolean
        subtle?: boolean
        fillWidth?: boolean
        failed?: boolean
    }) => {
        const [searchRef, setSearchRef] = useState(null)
        const [textRef, setTextRef] = useState(null)
        const isFocused = useIsFocused(searchRef)
        const textSize = useResize(textRef)
        const [error, setError] = useState(false)

        const [internalValue, setInternalValue] = useState(
            value !== undefined ? value : ''
        )
        useEffect(() => {
            setInternalValue(value !== undefined ? value : '')
        }, [value])

        useEffect(() => {
            if (isFocused === true && openInput) {
                openInput()
            } else if (isFocused === false && closeInput) {
                closeInput()
            }
        }, [isFocused, value])

        return (
            <div
                className="weblens-input-wrapper"
                style={{ height: height, minWidth: height }}
                data-value={internalValue}
                data-minimize={minimize}
                data-subtle={subtle}
                data-failed={failed || error}
                data-fill-width={fillWidth}
                onClick={(e) => {
                    e.stopPropagation()

                    if (searchRef) {
                        searchRef.focus()
                    }
                }}
                onDoubleClick={(e) => {
                    e.stopPropagation()
                }}
                onMouseDown={(e) => {
                    e.stopPropagation()
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
                {Icon && <Icon className="w-max h-max shrink-0" />}
                <p ref={setTextRef} className="weblens-input-text">
                    {internalValue}
                </p>
                <input
                    className="weblens-input select-none"
                    name={'input'}
                    style={{ width: !fillWidth ? textSize.width : '' }}
                    ref={setSearchRef}
                    autoFocus={autoFocus}
                    value={internalValue}
                    placeholder={placeholder}
                    type={password ? 'password' : ''}
                    onKeyDown={(e) => {
                        if (e.key === 'Escape') {
                            e.stopPropagation()
                            e.preventDefault()
                            if (searchRef) {
                                searchRef.blur()
                                setSearchRef(null)
                            }

                            return
                        } else if (e.key === 'Enter') {
                            if (onComplete && isFocused === true) {
                                e.stopPropagation()
                                onComplete(internalValue)
                                    .then(() => {
                                        if (closeInput) {
                                            closeInput()
                                        }
                                    })
                                    .catch((err) => {
                                        console.error(err)
                                        setError(true)
                                        setTimeout(() => setError(false), 2000)
                                    })
                            } else {
                                return
                            }
                        }
                        if (isFocused && !ignoreKeys.includes(e.key)) {
                            e.stopPropagation()
                        }
                        if (
                            stealFocus &&
                            !e.metaKey &&
                            ((e.which >= 65 && e.which <= 90) ||
                                e.key === 'Backspace')
                        ) {
                            e.stopPropagation()
                        }
                    }}
                    onTouchStart={(e) => {
                        e.preventDefault()
                    }}
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
                                if (onComplete) {
                                    onComplete(internalValue)
                                }
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
        } else if (prev.placeholder !== next.placeholder) {
            return false
        }
        return true
    }
)

export default WeblensInput
