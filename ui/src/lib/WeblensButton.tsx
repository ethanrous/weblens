import { useResize } from '@weblens/lib/hooks'
import { ErrorHandler } from '@weblens/types/Types'
import React, { CSSProperties, useEffect, useState } from 'react'

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
    setFail: (fail: boolean) => void
) => {
    if (!handler) {
        return
    }
    // Don't flash loading if handler returns instantly
    const tm = setTimeout(() => {
        setLoading(true)
    }, 150)

    try {
        const res = await handler(e)

        clearTimeout(tm)
        setLoading(false)
        if (res && showSuccess) {
            setSuccess(true)
            setTimeout(() => setSuccess(false), 2000)
        } else if (res === false && showSuccess) {
            setFail(true)
            setTimeout(() => setFail(false), 2000)
        }
    } catch (e) {
        clearTimeout(tm)
        setLoading(false)
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
    const [success, setSuccess] = useState(false)
    const [fail, setFail] = useState(false)
    const [loading, setLoading] = useState(false)

    // TODO: implement these
    console.debug(success, fail, loading)

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
                disabled={disabled}
                onClick={(e) => {
                    handleButtonEvent(
                        e,
                        onClick,
                        showSuccess,
                        setLoading,
                        setSuccess,
                        setFail
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
                        'text-button-text flex items-center justify-center data-[size=default]:h-6 data-[size=small]:text-sm data-[size=tiny]:text-xs'
                    }
                    data-size={size}
                >
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
                </span>
            </button>
        </div>
    )
}

// function WeblensButton({
//     label,
//     tooltip = '',
//     showSuccess = true,
//     toggleOn = undefined,
//     subtle = false,
//     allowRepeat = false,
//     centerContent = false,
//     disabled = false,
//     danger = false,
//     doSuper = false,
//     labelOnHover = false,
//     Left = null,
//     Right = null,
//     fillWidth = false,
//     allowShrink = true,
//     onClick,
//     type = 'button',
//     squareSize = 40,
//     float = false,
//     requireConfirm = false,
//
//     onMouseUp,
//     onMouseOver,
//     onMouseLeave,
//     style,
//     className,
//     setButtonRef = () => {},
// }: ButtonProps) {
//     const [success, setSuccess] = useState(false)
//     const [fail, setFail] = useState(false)
//     const [loading, setLoading] = useState(false)
//     const [textWidth, setTextWidth] = useState<number>(null)
//     const [hovering, setHovering] = useState(false)
//     const [confirming, setConfirming] = useState(false)
//
//     const [sizeRef, setSizeRef] = useState<HTMLDivElement>(null)
//     const buttonSize = useResize(sizeRef)
//
//     const iconSize = squareSize * 0.6
//
//     const targetWidth = useMemo(() => {
//         if (fillWidth) {
//             return '100%'
//         }
//         if (!label) {
//             return squareSize
//         }
//
//         if (!textWidth || textWidth === -1) {
//             return 'max-content'
//         }
//
//         let returnWidth = 16
//
//         if (buttonSize.width > textWidth + squareSize || (!Left && !Right)) {
//             returnWidth += textWidth
//         }
//
//         if (Left) {
//             returnWidth += squareSize
//         }
//         if (Right) {
//             returnWidth += squareSize
//         }
//
//         return returnWidth
//     }, [buttonSize, fillWidth, squareSize, label, textWidth, toggleOn])
//
//     const maxWidth = useMemo(() => {
//         if (fillWidth) {
//             return ''
//         }
//         if (!label) {
//             return squareSize
//         }
//
//         if (labelOnHover && !hovering) {
//             return squareSize
//         }
//
//         // return textWidth + squareSize + 16 + 1000
//         return 'max-content'
//     }, [fillWidth, squareSize, label, textWidth, hovering])
//
//     return (
//         <div
//             ref={setSizeRef}
//             className={
//                 buttonStyle.weblensButtonWrapper +
//                 ' ' +
//                 className +
//                 ' ' +
//                 'group/button'
//             }
//             data-fill-width={fillWidth}
//             data-text-on-hover={labelOnHover}
//             style={{
//                 maxHeight: squareSize,
//                 minWidth: squareSize,
//                 width: targetWidth,
//                 // maxWidth: maxWidth,
//                 height: squareSize,
//                 flexShrink: Number(allowShrink),
//             }}
//         >
//             {(tooltip || confirming) && (
//                 <div
//                     className="pointer-events-none absolute w-max"
//                     style={{
//                         transform: `translateY(${squareSize / 2 + 20}px)`,
//                     }}
//                 >
//                     <span
//                         className="relative z-50 mt-1 truncate text-nowrap rounded-md border border-color-border-primary bg-background-secondary p-1 text-color-text-primary opacity-0 shadow-lg transition group-hover/button:opacity-100"
//                         style={{
//                             width: buttonSize.width * 2,
//                             maxWidth: 'max-content',
//                         }}
//                     >
//                         {confirming ? 'Really?' : tooltip}
//                     </span>
//                 </div>
//             )}
//             <button
//                 className={buttonStyle.weblensButton}
//                 ref={setButtonRef}
//                 data-disabled={disabled}
//                 data-toggled={toggleOn}
//                 data-repeat={allowRepeat}
//                 data-success={success}
//                 data-fail={fail}
//                 data-loading={loading}
//                 data-fill-width={fillWidth}
//                 data-center={centerContent}
//                 data-subtle={subtle}
//                 data-super={doSuper}
//                 data-danger={danger}
//                 data-float={float}
//                 type={type}
//                 style={{ ...style, width: targetWidth }}
//                 onClick={(e) => {
//                     if (!requireConfirm || confirming) {
//                         handleButtonEvent(
//                             e,
//                             onClick,
//                             showSuccess,
//                             setLoading,
//                             setSuccess,
//                             setFail
//                         ).catch(ErrorHandler)
//                         setConfirming(false)
//                     } else if (requireConfirm) {
//                         setConfirming(true)
//                         setTimeout(() => setConfirming(false), 2000)
//                     }
//                 }}
//                 onMouseUp={(e) => {
//                     handleButtonEvent(
//                         e,
//                         onMouseUp,
//                         showSuccess,
//                         setLoading,
//                         setSuccess,
//                         setFail
//                     ).catch(ErrorHandler)
//                 }}
//                 onMouseOver={(e) => {
//                     setHovering(true)
//                     {
//                         handleButtonEvent(
//                             e,
//                             onMouseOver,
//                             showSuccess,
//                             setLoading,
//                             setSuccess,
//                             setFail
//                         ).catch(ErrorHandler)
//                     }
//                 }}
//                 onMouseLeave={(e) => {
//                     setTimeout(() => setHovering(false), 200)
//                     handleButtonEvent(
//                         e,
//                         onMouseLeave,
//                         showSuccess,
//                         setLoading,
//                         setSuccess,
//                         setFail
//                     ).catch(ErrorHandler)
//                 }}
//             >
//                 {success && showSuccess && (
//                     <div
//                         className={buttonStyle.buttonContent + ' absolute'}
//                         data-center={true}
//                     >
//                         <IconCheck />
//                     </div>
//                 )}
//                 {fail && showSuccess && (
//                     <div
//                         className={buttonStyle.buttonContent + ' absolute'}
//                         data-center={true}
//                     >
//                         <IconX color="white" />
//                     </div>
//                 )}
//                 {loading && showSuccess && (
//                     <div
//                         className={
//                             buttonStyle.buttonContent + ' absolute h-full'
//                         }
//                         data-center={true}
//                     >
//                         <Loader size={squareSize / 2} color={'white'} />
//                     </div>
//                 )}
//                 {/* {!loading && !success && !fail && ( */}
//                 <ButtonContent
//                     label={label}
//                     Left={confirming ? IconQuestionMark : Left}
//                     Right={Right}
//                     staticTextWidth={textWidth}
//                     setTextWidth={setTextWidth}
//                     buttonWidth={
//                         hovering && labelOnHover
//                             ? textWidth + squareSize
//                             : buttonSize.width
//                     }
//                     iconSize={iconSize}
//                     centerContent={centerContent}
//                     hidden={success || fail || loading}
//                     labelOnHover={labelOnHover}
//                 />
//                 {/* )} */}
//             </button>
//         </div>
//     )
// }

export default WeblensButton
