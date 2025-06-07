import WeblensLoader from '@weblens/components/Loading'
import { useResize } from '@weblens/lib/hooks'
import { ErrorHandler } from '@weblens/types/Types'
import React, { CSSProperties, useEffect, useRef, useState } from 'react'

import LoaderDots from './LoaderDots'
import { ButtonActionHandler, ButtonProps } from './buttonTypes'

const handleButtonEvent = async (
    e: React.MouseEvent<HTMLElement, MouseEvent>,
    handler: ButtonActionHandler,
    showSuccess: boolean,
    setLoading: (loading: boolean) => void,
    setSuccess: (success: boolean) => void,
    setFail: (fail: boolean) => void,
    doLoading: boolean = false
) => {
    if (!handler) {
        return
    }

    let tm: NodeJS.Timeout
    if (doLoading) {
        // Don't flash loading if handler returns instantly
        tm = setTimeout(() => {
            setLoading(true)
        }, 150)
    }

    try {
        const res = await handler(e)

        if (doLoading) {
            clearTimeout(tm)
            setLoading(false)
        }
        if (res && showSuccess) {
            setSuccess(true)
            setTimeout(() => setSuccess(false), 2000)
        } else if (res === false && showSuccess) {
            setFail(true)
            setTimeout(() => setFail(false), 2000)
        }
    } catch (e) {
        if (doLoading) {
            clearTimeout(tm)
            setLoading(false)
        }
        ErrorHandler(Error(String(e)))
        if (showSuccess) {
            setSuccess(false)
            setFail(true)
            setTimeout(() => setFail(false), 2000)
        }
    }
}

function WeblensButton({
    ref,
    label,
    showSuccess = true,
    disabled = false,
    danger = false,
    Left = null,
    Right = null,
    fillWidth = false,

    tooltip,

    flavor = 'default',
    size = 'default',
    centerContent = false,
    className,

    onClick,
    onContextMenu,
    onMouseUp,
    onMouseLeave,
}: ButtonProps) {
    const [, setSuccess] = useState(false)
    const [, setFail] = useState(false)
    const [loading, setLoading] = useState(false)

    // TODO: implement these
    // console.debug(success, fail, loading)

    const [showLabel, setShowLabel] = useState(true)

    const buttonRef = useRef<HTMLButtonElement>(null)
    const buttonSize = useResize(buttonRef)

    useEffect(() => {
        if (Left === null && Right === null) {
            return
        }

        if (buttonSize.width !== -1 && buttonSize.width < 60) {
            setShowLabel(false)
        } else if (buttonSize.width >= 60) {
            setShowLabel(true)
        }
    }, [buttonSize, Left, Right])

    useEffect(() => {
        setShowLabel(true)
    }, [label])

    let buttonColor = '--color-button-primary'
    let buttonHoverColor = '--color-button-primary-hover'
    let buttonTextColor = '--color-button-text-primary'
    if (danger) {
        buttonColor = '--color-button-danger'
        buttonHoverColor = '--color-button-danger-hover'
    }

    switch (flavor) {
        case 'default':
            buttonTextColor = '--color-text-near-white'
            break
        case 'outline':
            buttonTextColor = '--color-text-primary'
            break
        case 'light':
            break
    }

    let iconSize = 24
    let buttonSpacing = '0.5rem'
    switch (size) {
        case 'default':
            break
        case 'small':
            iconSize = 20
            buttonSpacing = '0.3rem'
            break
        case 'tiny':
            iconSize = 16
            buttonSpacing = '0.2rem'
            break
        case 'jumbo':
            iconSize = 48
            buttonSpacing = '0.75rem'
            break
    }

    let tooltipClass = ''
    let doAutoTooltipPosition = false
    if (typeof tooltip === 'object') {
        if (tooltip.className) {
            tooltipClass = tooltip.className
        }
        // if (tooltip.position === 'right') {
        //     tooltipClass += ' right-[anchor(left)]'
        // }
        doAutoTooltipPosition = tooltip.position == 'auto'
    }

    return (
        <div
            className="bg-background-primary group relative flex h-max min-h-0 w-max justify-center rounded [--tooltip-left:auto] [--tooltip-right:auto] odd:[--tooltip-left:left] even:[--tooltip-right:right] data-fill-width:w-full"
            data-fill-width={fillWidth ? true : null}
            ref={ref}
        >
            <button
                ref={buttonRef}
                style={
                    {
                        '--color-button': `var(${buttonColor})`,
                        '--color-button-hover': `var(${buttonHoverColor})`,
                        '--color-button-text': `var(${buttonTextColor})`,
                        '--wl-button-spacing': buttonSpacing,
                        display: 'flex',
                        alignItems: centerContent ? 'center' : 'flex-start',
                    } as CSSProperties
                }
                data-fill-width={fillWidth}
                className={className}
                data-flavor={flavor}
                disabled={disabled || loading}
                onClick={(e) => {
                    handleButtonEvent(
                        e,
                        onClick,
                        showSuccess,
                        setLoading,
                        setSuccess,
                        setFail,
                        true
                    ).catch(ErrorHandler)
                }}
                onContextMenu={(e) => {
                    handleButtonEvent(
                        e,
                        onContextMenu,
                        showSuccess,
                        setLoading,
                        setSuccess,
                        setFail,
                        true
                    ).catch(ErrorHandler)
                }}
                onMouseUp={(e) => {
                    handleButtonEvent(
                        e,
                        onMouseUp,
                        showSuccess,
                        setLoading,
                        setSuccess,
                        setFail
                    ).catch(ErrorHandler)
                }}
                onMouseLeave={(e) => {
                    handleButtonEvent(
                        e,
                        onMouseLeave,
                        showSuccess,
                        setLoading,
                        setSuccess,
                        setFail
                    ).catch(ErrorHandler)
                }}
            >
                <span
                    className={
                        'flex items-center justify-center text-inherit data-[size=default]:h-6 data-[size=jumbo]:h-10 data-[size=jumbo]:text-2xl data-[size=small]:h-6 data-[size=small]:text-sm data-[size=tiny]:h-5 data-[size=tiny]:text-xs'
                    }
                    data-size={size}
                >
                    {loading && (
                        <div className="flex h-6 w-6 items-center justify-center">
                            <WeblensLoader size={16} />
                        </div>
                    )}
                    {!loading && (
                        <>
                            {Left && (
                                <span className="me-(--wl-button-spacing) text-inherit only:me-0">
                                    <Left size={iconSize} />
                                </span>
                            )}
                            {label && showLabel && (
                                <span className="text-[length:inherit] leading-none text-nowrap text-inherit">
                                    {label}
                                </span>
                            )}

                            {Right && (
                                <span className="ms-(--wl-button-spacing) text-inherit only:ms-0">
                                    <Right size={iconSize} />
                                </span>
                            )}
                        </>
                    )}
                </span>
            </button>
            {tooltip && (
                <div
                    className={[
                        'bg-background-secondary text-color-text-primary pointer-events-none absolute top-[100%] z-50 hidden h-max max-w-max min-w-max items-center justify-center rounded-md border p-1 opacity-100 shadow-lg transition group-hover:flex group-hover:opacity-100',
                        doAutoTooltipPosition
                            ? 'right-[anchor(var(--tooltip-right),0px)] left-[anchor(var(--tooltip-left),0px)]'
                            : '',
                        tooltipClass,
                    ].join(' ')}
                >
                    <span className="text-nowrap">
                        {typeof tooltip === 'string'
                            ? tooltip
                            : tooltip.content}
                    </span>
                </div>
            )}
        </div>
    )
}

export default WeblensButton
