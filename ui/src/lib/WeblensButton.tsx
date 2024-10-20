import { Loader } from '@mantine/core'
import { IconCheck, IconQuestionMark, IconX } from '@tabler/icons-react'
import { useResize } from '@weblens/components/hooks'
import React, {
    CSSProperties,
    memo,
    ReactNode,
    useEffect,
    useMemo,
    useState,
} from 'react'

import '@weblens/lib/weblensButton.scss'

type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | boolean | Promise<void | boolean | Response>

type buttonProps = {
    label?: string
    tooltip?: string
    showSuccess?: boolean
    toggleOn?: boolean
    subtle?: boolean
    allowRepeat?: boolean
    centerContent?: boolean
    danger?: boolean
    disabled?: boolean
    doSuper?: boolean
    labelOnHover?: boolean
    fillWidth?: boolean
    allowShrink?: boolean
    float?: boolean
    requireConfirm?: boolean
    Left?: (p: any) => ReactNode
    Right?: (p: any) => ReactNode

    // Style
    squareSize?: number
    fontSize?: string
    textMin?: number

    onClick?: ButtonActionHandler
    onMouseUp?: ButtonActionHandler
    onMouseOver?: ButtonActionHandler
    onMouseLeave?: ButtonActionHandler
    style?: CSSProperties
    setButtonRef?: (ref: HTMLDivElement) => void
}

const ButtonContent = memo(
    ({
        label,
        Left,
        Right,
        setTextWidth,
        buttonWidth,
        iconSize,
        centerContent,
        hidden,
        labelOnHover,
    }: {
        label: string
        Left: (p: any) => ReactNode
        Right: (p: any) => ReactNode
        setTextWidth: (w: number) => void
        buttonWidth: number
        iconSize: number
        centerContent: boolean
        hidden: boolean
        labelOnHover: boolean
    }) => {
        const [textRef, setTextRef] = useState<HTMLParagraphElement>()
        const { width: textWidth } = useResize(textRef)

        useEffect(() => {
            if (textWidth !== -1) {
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
                buttonWidth === 0 ||
                buttonWidth > textWidth
            )

            // return !(
            //     (!Boolean(label) ||
            //         ((Boolean(Left) || Boolean(Right)) && buttonWidth < iconSize + textWidth && buttonWidth !== 0))
            // && !(buttonWidth > textWidth) );
        }, [buttonWidth, textWidth])

        if (!iconSize) {
            iconSize = 24
        }

        return (
            <div
                className="button-content"
                data-center={centerContent || !showText}
                data-hidden={hidden}
                data-has-icon={Boolean(Left || Right)}
            >
                <div
                    className="button-icon-box"
                    data-has-icon={Boolean(Left)}
                    data-has-text={showText}
                    style={{
                        height: iconSize,
                        width: iconSize,
                    }}
                >
                    {Left && <Left className="button-icon" />}
                </div>
                <div
                    className="button-text-box"
                    data-show-text={showText}
                    data-center={centerContent}
                    data-hover-only={labelOnHover}
                >
                    <p
                        className="button-text"
                        ref={setTextRef}
                        data-show-text={showText}
                    >
                        {label}
                    </p>
                </div>

                <div
                    className="button-icon-box"
                    data-has-icon={Boolean(Right)}
                    data-has-text={showText}
                    data-icon-side={'right'}
                    style={{
                        height: iconSize,
                        width: iconSize,
                    }}
                >
                    {Right && <Right className="button-icon" />}
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.buttonWidth !== next.buttonWidth) {
            return false
        } else if (prev.label !== next.label) {
            return false
        } else if (prev.Left !== next.Left) {
            return false
        } else if (prev.hidden !== next.hidden) {
            return false
        } else if (prev.iconSize !== next.iconSize) {
            return false
        }
        return true
    }
)

const handleButtonEvent = async (
    e: React.MouseEvent<HTMLElement, MouseEvent>,
    handler: ButtonActionHandler,
    showSuccess: boolean,
    setLoading,
    setSuccess,
    setFail
) => {
    if (!handler) {
        return
    }
    const tm = setTimeout(() => {
        setLoading(true)
    }, 150)
    try {
        // Don't flash loading if handler returns instantly
        const res = await handler(e)

        clearTimeout(tm)
        setLoading(false)
        if (res === true && showSuccess) {
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

const WeblensButton = memo(
    ({
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
    }: buttonProps) => {
        const [success, setSuccess] = useState(false)
        const [fail, setFail] = useState(false)
        const [loading, setLoading] = useState(false)
        const [textWidth, setTextWidth] = useState(null)
        const [hovering, setHovering] = useState(false)
        const [confirming, setConfirming] = useState(false)

        const [sizeRef, setSizeRef] = useState<HTMLDivElement>(null)
        const buttonSize = useResize(sizeRef)

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

            let returnWidth = textWidth + 16

            if (Left) {
                returnWidth += squareSize
            }
            if (Right) {
                returnWidth += squareSize
            }

            return returnWidth
        }, [fillWidth, squareSize, label, textWidth, toggleOn])

        const maxWidth = useMemo(() => {
            if (fillWidth) {
                return ''
            }
            if (!label) {
                return squareSize
            }
            if (hovering && labelOnHover) {
                return textWidth + squareSize + 16
            } else if (labelOnHover) {
                return squareSize
            }

            return 'max-content'
        }, [fillWidth, squareSize, label, textWidth, hovering])

        return (
            <div
                ref={setSizeRef}
                className="weblens-button-wrapper"
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
                        className="button-tooltip"
                        style={{
                            transform: `translateY(${squareSize / 2 + 25}px)`,
                        }}
                    >
                        <p className="text-white text-nowrap">
                            {confirming ? 'Really?' : tooltip}
                        </p>
                    </div>
                )}
                <div
                    className="weblens-button"
                    ref={setButtonRef}
                    data-disabled={disabled}
                    data-toggled={toggleOn}
                    data-repeat={allowRepeat}
                    data-success={success}
                    data-fill-width={fillWidth}
                    data-center={centerContent}
                    data-fail={fail}
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
                            )
                            setConfirming(false)
                        } else if (requireConfirm) {
                            setConfirming(true)
                            setTimeout(() => setConfirming(false), 2000)
                        }
                    }}
                    onMouseUp={(e) =>
                        handleButtonEvent(
                            e,
                            onMouseUp,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        )
                    }
                    onMouseOver={(e) => {
                        setHovering(true)
                        handleButtonEvent(
                            e,
                            onMouseOver,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        )
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
                        )
                    }}
                >
                    {success && showSuccess && (
                        <div
                            className="button-content absolute"
                            data-center={true}
                        >
                            <IconCheck />
                        </div>
                    )}
                    {fail && showSuccess && (
                        <div
                            className="button-content absolute"
                            data-center={true}
                        >
                            <IconX />
                        </div>
                    )}
                    {loading && showSuccess && (
                        <div
                            className="button-content h-full"
                            data-center={true}
                        >
                            <Loader size={squareSize / 2} color={'white'} />
                        </div>
                    )}
                    {!loading && !success && !fail && (
                        <ButtonContent
                            label={label}
                            Left={confirming ? IconQuestionMark : Left}
                            Right={Right}
                            setTextWidth={setTextWidth}
                            buttonWidth={
                                hovering && labelOnHover
                                    ? textWidth + squareSize
                                    : buttonSize.width
                            }
                            iconSize={squareSize * 0.6}
                            centerContent={centerContent}
                            hidden={success || fail || loading}
                            labelOnHover={labelOnHover}
                        />
                    )}
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.toggleOn !== next.toggleOn) {
            return false
        } else if (prev.label !== next.label) {
            return false
        } else if (prev.disabled !== next.disabled) {
            return false
        } else if (prev.onClick !== next.onClick) {
            return false
        } else if (prev.onMouseUp !== next.onMouseUp) {
            return false
        } else if (prev.onMouseOver !== next.onMouseOver) {
            return false
        } else if (prev.squareSize !== next.squareSize) {
            return false
        } else if (prev.style !== next.style) {
            return false
        } else if (prev.Left !== next.Left) {
            return false
        }
        return true
    }
)

export default WeblensButton
