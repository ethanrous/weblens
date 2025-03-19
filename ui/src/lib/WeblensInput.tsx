import { Icon } from '@tabler/icons-react'
import { useIsFocused } from '@weblens/lib/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { useEffect, useState } from 'react'

import { ButtonActionPromiseReturn, ButtonIcon } from './buttonTypes'
import inputStyle from './weblensInput.module.scss'

function WeblensInput({
    onComplete,
    value,
    valueCallback,
    Icon,
    buttonIcon,
    squareSize,
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
    valid,
    autoComplete = 'off',
    className = '',
    disabled = false,
}: {
    onComplete?: (val: string) => ButtonActionPromiseReturn
    value?: string
    valueCallback?: (v: string) => void
    Icon?: Icon
    buttonIcon?: ButtonIcon
    squareSize?: number
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
    valid?: boolean
    autoComplete?: string
    className?: string
    disabled?: boolean
}) {
    const [searchRef, setSearchRef] = useState<HTMLInputElement>(null)
    const isFocused = useIsFocused(searchRef)
    const [error, setError] = useState(false)
    const [wrapperRef, setWrapperRef] = useState<HTMLDivElement>(null)

    const [internalValue, setInternalValue] = useState(
        value !== undefined ? value : ''
    )
    useEffect(() => {
        setInternalValue(value !== undefined ? value : '')
    }, [value])

    useEffect(() => {
        if (isFocused === true && openInput) {
            openInput()
        }
    }, [isFocused, value])

    return (
        <div
            className={inputStyle.weblensInputWrapper + ' ' + className}
            ref={setWrapperRef}
            style={{
                height: squareSize,
                maxHeight: squareSize,
                minWidth: squareSize,
            }}
            data-value={internalValue}
            data-minimize={minimize}
            data-subtle={subtle}
            data-failed={valid === false || error}
            data-valid={valid === true && !error}
            data-fill-width={fillWidth}
            data-disabled={disabled}
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
                if (closeInput && !wrapperRef.contains(e.relatedTarget)) {
                    e.stopPropagation()
                    closeInput()
                }
            }}
        >
            {Icon && <Icon className="h-max w-max shrink-0" />}
            <input
                className="h-full w-full"
                name={'input'}
                style={{ width: fillWidth ? '' : 'max-content' }}
                ref={setSearchRef}
                autoFocus={autoFocus}
                value={internalValue}
                placeholder={placeholder}
                type={password ? 'password' : 'text'}
                autoComplete={autoComplete}
                disabled={disabled}
                onKeyDown={(e) => {
                    if (e.ctrlKey) {
                        return
                    }
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
                            if (
                                valid === false ||
                                error ||
                                internalValue === ''
                            ) {
                                return
                            }
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
                        squareSize={squareSize ? squareSize * 0.75 : 40}
                        Left={buttonIcon}
                        disabled={
                            valid === false || error || internalValue === ''
                        }
                        onClick={(e) => {
                            e.stopPropagation()
                            if (onComplete) {
                                return onComplete(internalValue)
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
}

export default WeblensInput
