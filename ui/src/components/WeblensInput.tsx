import { memo, ReactNode, useEffect, useState } from 'react'
import { useIsFocused, useKeyDown, useResize } from './hooks'
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
        password = false,
        minimize = false,
        subtle = false,
        fillWidth = true,
        failed = false,
    }: {
        onComplete?: (v: string) => void
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
        password?: boolean
        minimize?: boolean
        subtle?: boolean
        fillWidth?: boolean
        failed?: boolean
    }) => {
        const [searchRef, setSearchRef] = useState(null)
        const [textRef, setTextRef] = useState(null)
        const [editing, setEditing] = useState(autoFocus)
        const isFocused = useIsFocused(searchRef)
        const textSize = useResize(textRef)

        const [internalValue, setInternalValue] = useState(
            value !== undefined ? value : ''
        )
        useEffect(() => {
            setInternalValue(value !== undefined ? value : '')
        }, [value])

        useEffect(() => {
            if (isFocused === true && openInput) {
                openInput()
                setEditing(true)
            } else if (isFocused === false && closeInput) {
                // closeInput();
                // setEditing(false);
            }
        }, [isFocused, value])

        useKeyDown(
            'Enter',
            () => {
                if (onComplete && isFocused === true) {
                    onComplete(internalValue)
                    if (closeInput) {
                        closeInput()
                    }
                }
            },
            !isFocused
        )

        useKeyDown('Escape', (e) => {
            e.stopPropagation()
            e.preventDefault()
            if (searchRef) {
                searchRef.blur()
                setSearchRef(null)
            }
            setEditing(false)
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
                setEditing(true)
            }
        )

        return (
            <div
                className="weblens-input-wrapper"
                style={{ height: height, minWidth: height }}
                data-value={internalValue}
                data-minimize={minimize}
                data-subtle={subtle}
                data-failed={failed}
                data-fill-width={fillWidth}
                onClick={(e) => {
                    e.stopPropagation()
                    setEditing(true)
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
