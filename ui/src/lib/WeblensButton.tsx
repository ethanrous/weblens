import { useResize } from '@weblens/components/hooks'
import { ErrorHandler } from '@weblens/types/Types'
import React, { CSSProperties, useEffect, useState } from 'react'

import { ButtonActionHandler, buttonProps as ButtonProps } from './buttonTypes'

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

    let buttonColor = '--wl-button-color-primary'
    let buttonHoverColor = '--wl-button-color-primary-hover'
    let buttonTextColor = '--wl-button-text-color-primary'
    if (danger) {
        buttonColor = '--wl-button-color-danger'
        buttonHoverColor = '--wl-button-color-danger-hover'
    }

    switch (flavor) {
        case 'default':
            buttonTextColor = '--wl-text-color-near-white'
            break
        case 'outline':
            buttonTextColor = '--wl-text-color-primary'
            break
        case 'light':
            break
    }

    let sizeClass = ''
    let iconSize = 24
    let buttonSpacing = '0.5rem'
    switch (size) {
        case 'default':
            sizeClass = 'h-max'
            break
        case 'small':
            sizeClass = 'text-s h-5'
            iconSize = 20
            buttonSpacing = '0.3rem'
            break
        case 'tiny':
            sizeClass = 'text-xs h-4'
            iconSize = 16
            buttonSpacing = '0.2rem'
            break
        case 'jumbo':
            sizeClass = 'text-xl'
            iconSize = 48
            buttonSpacing = '0.75rem'
            break
    }

    return (
        <button
            ref={setButtonRef}
            style={
                {
                    '--wl-button-color': `var(${buttonColor})`,
                    '--wl-button-hover-color': `var(${buttonHoverColor})`,
                    '--wl-button-text-color': `var(${buttonTextColor})`,
                    '--wl-button-spacing': buttonSpacing,
                } as CSSProperties
            }
            className={className}
            data-fill-width={fillWidth}
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
                    'flex items-center justify-center text-wl-button-text-color' +
                    sizeClass
                }
            >
                {Left && (
                    <span className="me-[--wl-button-spacing] text-inherit only:me-0">
                        <Left size={iconSize} />
                    </span>
                )}
                {label && showLabel && (
                    <span className="text-nowrap text-[length:inherit] text-inherit">
                        {label}
                    </span>
                )}

                {Right && (
                    <span className="ms-[--wl-button-spacing] text-inherit only:ms-0">
                        <Right size={iconSize} />
                    </span>
                )}
            </span>
        </button>
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
//                         className="relative z-50 mt-1 truncate text-nowrap rounded-md border border-wl-border-color-primary bg-wl-background-color-secondary p-1 text-wl-text-color-primary opacity-0 shadow-lg transition group-hover/button:opacity-100"
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
