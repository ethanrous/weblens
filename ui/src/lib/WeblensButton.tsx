import { useResize } from '@weblens/lib/hooks'
import { ErrorHandler } from '@weblens/types/Types'
import React, { CSSProperties, useEffect, useState } from 'react'

import LoaderDots from './LoaderDots'
import { ButtonActionHandler, ButtonProps } from './buttonTypes'

// function ButtonContent({
//     label,
//     Left,
//     Right,
//     staticTextWidth,
//     setTextWidth,
//     buttonWidth,
//     iconSize,
//     centerContent,
//     hidden,
//     labelOnHover,
// }: ButtonContentProps) {
//     const [textRef, setTextRef] = useState<HTMLParagraphElement>()
//     const { width: textWidth } = useResize(textRef)
//
//     useEffect(() => {
//         if (textWidth !== -1 && !staticTextWidth) {
//             setTextWidth(textWidth)
//         }
//     }, [textWidth])
//
//     const showText = useMemo(() => {
//         if (buttonWidth === -1 || textWidth === -1) {
//             return true
//         }
//         if (!label) {
//             return false
//         } else if (!Left && !Right) {
//             return true
//         }
//
//         return (
//             (Boolean(label) && !Left && !Right) ||
//             buttonWidth >= iconSize + textWidth ||
//             buttonWidth === 0
//         )
//     }, [buttonWidth, textWidth])
//
//     if (!iconSize) {
//         iconSize = 24
//     }
//
//     return (
//         <div
//             className={buttonStyle.buttonContent}
//             data-center={centerContent || !showText}
//             data-hidden={hidden}
//             data-has-icon={Boolean(Left || Right)}
//         >
//             <div
//                 className={buttonStyle.buttonIconBox}
//                 data-has-icon={Boolean(Left)}
//                 data-has-text={showText}
//                 style={{
//                     height: iconSize,
//                     width: iconSize,
//                 }}
//             >
//                 {Left && <Left className={buttonStyle.buttonIcon} />}
//             </div>
//             <div
//                 className={buttonStyle.buttonTextBox}
//                 data-show-text={showText}
//                 data-center={centerContent}
//                 data-hover-only={labelOnHover}
//             >
//                 <p
//                     className={buttonStyle.buttonText}
//                     ref={setTextRef}
//                     data-show-text={showText}
//                 >
//                     {label}
//                 </p>
//             </div>
//
//             <div
//                 className={buttonStyle.buttonIconBox}
//                 data-has-icon={Boolean(Right)}
//                 data-has-text={showText}
//                 data-icon-side={'right'}
//                 style={{
//                     height: iconSize,
//                     // width: iconSize,
//                 }}
//             >
//                 {Right && <Right className={buttonStyle.buttonIcon} />}
//             </div>
//         </div>
//     )
// }

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
    label,
    showSuccess = true,
    disabled = false,
    danger = false,
    Left = null,
    Right = null,
    fillWidth = false,
    onClick,

    flavor = 'default',
    size = 'default',

    onMouseUp,
    onMouseLeave,
    className,
}: ButtonProps) {
    const [, setSuccess] = useState(false)
    const [, setFail] = useState(false)
    const [loading, setLoading] = useState(false)

    // TODO: implement these
    // console.debug(success, fail, loading)

    const [showLabel, setShowLabel] = useState(true)
    const [buttonRef, setButtonRef] = useState<HTMLButtonElement>()

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
    }, [buttonSize])

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

    return (
        <div
            className="bg-background-primary flex h-max min-h-0 w-max rounded data-fill-width:w-full"
            data-fill-width={fillWidth ? true : null}
        >
            <button
                ref={setButtonRef}
                style={
                    {
                        '--color-button': `var(${buttonColor})`,
                        '--color-button-hover': `var(${buttonHoverColor})`,
                        '--color-button-text': `var(${buttonTextColor})`,
                        '--wl-button-spacing': buttonSpacing,
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
                        'text-button-text flex items-center justify-center data-[size=default]:h-6 data-[size=jumbo]:h-10 data-[size=jumbo]:text-2xl data-[size=small]:h-6 data-[size=small]:text-sm data-[size=tiny]:h-5 data-[size=tiny]:text-xs'
                    }
                    data-size={size}
                >
                    {loading && <LoaderDots />}
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
        </div>
    )
}

export default WeblensButton
