import { Loader } from '@mantine/core'
import { IconCheck, IconQuestionMark, IconX } from '@tabler/icons-react'
import { useResize } from '@weblens/components/hooks'
import { ErrorHandler } from '@weblens/types/Types'
import React, { useEffect, useMemo, useState } from 'react'

import {
    ButtonActionHandler,
    ButtonContentProps,
    buttonProps as ButtonProps,
} from './buttonTypes'
import buttonStyle from './weblensButton.module.scss'

function ButtonContent({
    label,
    Left,
    Right,
    staticTextWidth,
    setTextWidth,
    buttonWidth,
    iconSize,
    centerContent,
    hidden,
    labelOnHover,
}: ButtonContentProps) {
    const [textRef, setTextRef] = useState<HTMLParagraphElement>()
    const { width: textWidth } = useResize(textRef)

    useEffect(() => {
        if (textWidth !== -1 && !staticTextWidth) {
            setTextWidth(textWidth)
        }
    }, [textWidth])

    const showText = useMemo(() => {
        if (buttonWidth === -1 || textWidth === -1) {
            return true
        }
        if (!label) {
            return false
        } else if (!Left && !Right) {
            return true
        }

        return (
            (Boolean(label) && !Left && !Right) ||
            buttonWidth >= iconSize + textWidth ||
            buttonWidth === 0
        )
    }, [buttonWidth, textWidth])

    if (!iconSize) {
        iconSize = 24
    }

    return (
        <div
            className={buttonStyle['button-content']}
            data-center={centerContent || !showText}
            data-hidden={hidden}
            data-has-icon={Boolean(Left || Right)}
        >
            <div
                className={buttonStyle['button-icon-box']}
                data-has-icon={Boolean(Left)}
                data-has-text={showText}
                style={{
                    height: iconSize,
                    width: iconSize,
                }}
            >
                {Left && <Left className={buttonStyle['button-icon']} />}
            </div>
            <div
                className={buttonStyle['button-text-box']}
                data-show-text={showText}
                data-center={centerContent}
                data-hover-only={labelOnHover}
            >
                <p
                    className={buttonStyle['button-text']}
                    ref={setTextRef}
                    data-show-text={showText}
                >
                    {label}
                </p>
            </div>

            <div
                className={buttonStyle['button-icon-box']}
                data-has-icon={Boolean(Right)}
                data-has-text={showText}
                data-icon-side={'right'}
                style={{
                    height: iconSize,
                    // width: iconSize,
                }}
            >
                {Right && <Right className={buttonStyle['button-icon']} />}
            </div>
        </div>
    )
}

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
        console.error(e)
        if (showSuccess) {
            setSuccess(false)
            setFail(true)
            setTimeout(() => setFail(false), 2000)
        }
    }
}

function WeblensButton({
    label,
    tooltip = '',
    showSuccess = true,
    toggleOn = undefined,
    subtle = false,
    allowRepeat = false,
    centerContent = false,
    disabled = false,
    danger = false,
    doSuper = false,
    labelOnHover = false,
    Left = null,
    Right = null,
    fillWidth = false,
    allowShrink = true,
    onClick,
    squareSize = 40,
    float = false,
    requireConfirm = false,

    onMouseUp,
    onMouseOver,
    onMouseLeave,
    style,
    setButtonRef = () => {},
}: ButtonProps) {
    const [success, setSuccess] = useState(false)
    const [fail, setFail] = useState(false)
    const [loading, setLoading] = useState(false)
    const [textWidth, setTextWidth] = useState<number>(null)
    const [hovering, setHovering] = useState(false)
    const [confirming, setConfirming] = useState(false)

    const [sizeRef, setSizeRef] = useState<HTMLDivElement>(null)
    const buttonSize = useResize(sizeRef)

    const iconSize = squareSize * 0.6

    const targetWidth = useMemo(() => {
        if (fillWidth) {
            return '100%'
        }
        if (!label) {
            return squareSize
        }

        if (!textWidth || textWidth === -1) {
            return 'max-content'
        }

        let returnWidth = 16

        if (buttonSize.width > textWidth + squareSize || (!Left && !Right)) {
            returnWidth += textWidth
        }

        if (Left) {
            returnWidth += squareSize
        }
        if (Right) {
            returnWidth += squareSize
        }

        return returnWidth
    }, [buttonSize, fillWidth, squareSize, label, textWidth, toggleOn])

    const maxWidth = useMemo(() => {
        if (fillWidth) {
            return ''
        }
        if (!label) {
            return squareSize
        }

        if (labelOnHover && !hovering) {
            return squareSize
        }

        // return textWidth + squareSize + 16 + 1000
        return 'max-content'
    }, [fillWidth, squareSize, label, textWidth, hovering])

    return (
        <div
            ref={setSizeRef}
            className={buttonStyle['weblens-button-wrapper']}
            data-fill-width={fillWidth}
            data-text-on-hover={labelOnHover}
            style={{
                maxHeight: squareSize,
                minWidth: squareSize,
                width: targetWidth,
                maxWidth: maxWidth,
                height: squareSize,
                flexShrink: Number(allowShrink),
            }}
        >
            {(tooltip || confirming) && (
                <div
                    className={buttonStyle['button-tooltip']}
                    style={{
                        transform: `translateY(${squareSize / 2 + 20}px)`,
                    }}
                >
                    <p
                        className="flex text-white z-10 grow text-nowrap"
                        style={{
                            width: buttonSize.width * 2,
                            maxWidth: 'max-content',
                        }}
                    >
                        {confirming ? 'Really?' : tooltip}
                    </p>
                </div>
            )}
            <div
                className={buttonStyle['weblens-button']}
                ref={setButtonRef}
                data-disabled={disabled}
                data-toggled={toggleOn}
                data-repeat={allowRepeat}
                data-success={success}
                data-fail={fail}
                data-loading={loading}
                data-fill-width={fillWidth}
                data-center={centerContent}
                data-subtle={subtle}
                data-super={doSuper}
                data-danger={danger}
                data-float={float}
                style={{ ...style, width: targetWidth }}
                onClick={(e) => {
                    if (!requireConfirm || confirming) {
                        handleButtonEvent(
                            e,
                            onClick,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        ).catch(ErrorHandler)
                        setConfirming(false)
                    } else if (requireConfirm) {
                        setConfirming(true)
                        setTimeout(() => setConfirming(false), 2000)
                    }
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
                onMouseOver={(e) => {
                    setHovering(true)
                    {
                        handleButtonEvent(
                            e,
                            onMouseOver,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        ).catch(ErrorHandler)
                    }
                }}
                onMouseLeave={(e) => {
                    setTimeout(() => setHovering(false), 200)
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
                {success && showSuccess && (
                    <div
                        className={buttonStyle['button-content'] + ' absolute'}
                        data-center={true}
                    >
                        <IconCheck />
                    </div>
                )}
                {fail && showSuccess && (
                    <div
                        className={buttonStyle['button-content'] + ' absolute'}
                        data-center={true}
                    >
                        <IconX color="white" />
                    </div>
                )}
                {loading && showSuccess && (
                    <div
                        className={
                            buttonStyle['button-content'] + ' absolute h-full'
                        }
                        data-center={true}
                    >
                        <Loader size={squareSize / 2} color={'white'} />
                    </div>
                )}
                {/* {!loading && !success && !fail && ( */}
                <ButtonContent
                    label={label}
                    Left={confirming ? IconQuestionMark : Left}
                    Right={Right}
                    staticTextWidth={textWidth}
                    setTextWidth={setTextWidth}
                    buttonWidth={
                        hovering && labelOnHover
                            ? textWidth + squareSize
                            : buttonSize.width
                    }
                    iconSize={iconSize}
                    centerContent={centerContent}
                    hidden={success || fail || loading}
                    labelOnHover={labelOnHover}
                />
                {/* )} */}
            </div>
        </div>
    )
}

export default WeblensButton
