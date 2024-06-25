import { Loader, Text } from '@mantine/core'
import React, {
    CSSProperties,
    memo,
    ReactNode,
    useCallback,
    useEffect,
    useMemo,
    useState,
} from 'react'
import { IconCheck, IconX } from '@tabler/icons-react'
import { useResize } from './hooks'

import './weblensButton.scss'

type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | boolean | Promise<boolean>

type buttonProps = {
    label?: string
    postScript?: string
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
    setButtonRef?
}

const ButtonContent = memo(
    ({
        label,
        postScript,
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
        postScript: string
        Left: (p: any) => ReactNode
        Right: (p: any) => ReactNode
        setTextWidth: (w: number) => void
        buttonWidth: number
        iconSize: number
        centerContent: boolean
        hidden: boolean
        labelOnHover: boolean
    }) => {
        const [textRef, setTextRef] = useState(null)
        const { width: textWidth } = useResize(textRef)
        const textFits = useMemo(() => {
            return buttonWidth > textWidth
        }, [buttonWidth, textWidth])
        const [showText, setShowText] = useState(Boolean(label) && textFits)

        useEffect(() => {
            setTextWidth(textWidth)
            if (
                !Boolean(label) ||
                ((Boolean(Left) || Boolean(Right)) &&
                    buttonWidth < iconSize + textWidth &&
                    buttonWidth !== 0)
            ) {
                if (showText) {
                    setShowText(false)
                }
            } else if (!showText && textFits) {
                setShowText(true)
            }
        }, [buttonWidth, textWidth])

        if (!iconSize) {
            iconSize = 24
        }

        return (
            <div
                className="button-content"
                data-center={centerContent || (!showText && !labelOnHover)}
                data-hidden={hidden}
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
                >
                    <p
                        className="button-text"
                        ref={setTextRef}
                        data-show-text={showText}
                        data-hover-only={labelOnHover}
                    >
                        {label}
                    </p>

                    {postScript && (
                        <p className="font-light text-xs text-ellipsis text-nowrap overflow-visible select-none">
                            {postScript}
                        </p>
                    )}
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
            // console.log('BUTTON WIDTH')
            return false
        } else if (prev.label !== next.label) {
            // console.log('LABEL')
            return false
        } else if (prev.postScript !== next.postScript) {
            // console.log('POST SCRIPT')
            return false
        } else if (prev.Left !== next.Left) {
            // console.log('LEFT')
            return false
        } else if (prev.hidden !== next.hidden) {
            // console.log('HIDDEN')
            return false
        }
        return true
    }
)

// if (toggleOn !== undefined) {
//     onToggle(!toggleOn);
// } else {
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
        postScript,
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
        onClick,
        squareSize = 40,

        onMouseUp,
        onMouseOver,
        onMouseLeave,
        style,
        setButtonRef = (r) => {},
    }: buttonProps) => {
        const [success, setSuccess] = useState(false)
        const [fail, setFail] = useState(false)
        const [loading, setLoading] = useState(false)
        const [textWidth, setTextWidth] = useState(0)
        const [maxW, setMaxW] = useState(labelOnHover ? 40 : 400)

        const [sizeRef, setSizeRef]: [
            buttonRef: HTMLDivElement,
            setButtonRef: any,
        ] = useState(null)
        const buttonSize = useResize(sizeRef)

        const targetWidth = useMemo(() => {
            if (fillWidth) {
                return '100%'
            } else if (label) {
                return textWidth + squareSize + 16
            } else {
                return squareSize
            }
        }, [fillWidth, squareSize, label, textWidth])

        const hoverCallback = useCallback(() => {
            if (labelOnHover) {
                setMaxW(textWidth + squareSize + 16)
            }
        }, [textWidth, squareSize, setMaxW])

        const unHoverCallback = useCallback(() => {
            if (labelOnHover) {
                setMaxW(40)
            }
        }, [setMaxW])

        return (
            <div
                ref={setSizeRef}
                className="weblens-button-wrapper"
                data-fill-width={fillWidth}
                data-text-on-hover={labelOnHover}
                onMouseOver={hoverCallback}
                onMouseLeave={unHoverCallback}
                style={{
                    maxHeight: squareSize,
                    width: targetWidth,
                    maxWidth: maxW,
                    height: squareSize,
                }}
            >
                <div
                    ref={setButtonRef}
                    className={
                        toggleOn === undefined
                            ? 'weblens-button'
                            : 'weblens-toggle-button'
                    }
                    data-disabled={disabled}
                    data-toggled={!!toggleOn}
                    data-repeat={allowRepeat}
                    data-success={success}
                    data-fail={fail}
                    data-subtle={subtle}
                    data-super={doSuper}
                    data-danger={danger}
                    style={{ width: maxW, ...style }}
                    onClick={(e) =>
                        handleButtonEvent(
                            e,
                            onClick,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        )
                    }
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
                    onMouseOver={(e) =>
                        handleButtonEvent(
                            e,
                            onMouseOver,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        )
                    }
                    onMouseLeave={(e) =>
                        handleButtonEvent(
                            e,
                            onMouseLeave,
                            showSuccess,
                            setLoading,
                            setSuccess,
                            setFail
                        )
                    }
                >
                    <div className="flex w-full relative overflow-hidden">
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
                            <div className="button-content" data-center={true}>
                                <Loader size={'20px'} color={'white'} />
                            </div>
                        )}
                        <ButtonContent
                            label={label}
                            postScript={postScript}
                            Left={Left}
                            Right={Right}
                            setTextWidth={setTextWidth}
                            buttonWidth={buttonSize.width}
                            iconSize={squareSize * 0.6}
                            centerContent={centerContent}
                            hidden={success || fail || loading}
                            labelOnHover={labelOnHover}
                        />
                    </div>
                </div>
            </div>
        )
    },
    (prev, next) => {
        if (prev.toggleOn !== next.toggleOn) {
            // console.log(next.label, 'TOGGLE')
            return false
        } else if (prev.label !== next.label) {
            // console.log(next.label, 'LABEL')
            return false
        } else if (prev.disabled !== next.disabled) {
            // console.log(next.label, 'DISABLED')
            return false
        } else if (prev.onClick !== next.onClick) {
            // console.log(next.label, 'ONCLICK')
            return false
        } else if (prev.onMouseUp !== next.onMouseUp) {
            // console.log(next.label, 'MOUSEUP')
            return false
        } else if (prev.onMouseOver !== next.onMouseOver) {
            // console.log(next.label, 'MOUSEOVER')
            return false
        } else if (prev.postScript !== next.postScript) {
            // console.log(next.label, 'POSTSCRIPT')
            return false
        } else if (prev.squareSize !== next.squareSize) {
            // console.log(next.label, 'SQUARESIZE')
            return false
        } else if (prev.style !== next.style) {
            // console.log(next.label, 'STYLE')
            return false
        } else if (prev.Left !== next.Left) {
            // console.log(next.label, 'LEFT')
            return false
        }
        return true
    }
)

export const SelectIcon = ({
    size,
    expandSize,
    label,
    icon,
    selected,
    index,
    selectedIndex,
    onClick,
}: {
    size: number
    expandSize?: number
    label?: string
    icon: JSX.Element
    selected: boolean
    index?: number
    selectedIndex?: number
    onClick?
}) => {
    const [hover, setHover] = useState(false)
    const iconBoxStyle = useMemo(() => {
        return {
            height: size,
            width: size,
        }
    }, [size])
    return (
        <div>
            <div
                className="weblens-select-icon"
                data-selected={selected}
                style={{ height: size, width: size }}
                onClick={(e) => {
                    onClick(e)
                }}
                onMouseOver={(e) => setHover(true)}
                onMouseLeave={(e) => setHover(false)}
            >
                {icon}
            </div>
            {hover && expandSize && (
                <div
                    className="flex items-center left-0 top-0 z-10 bg-[#222222] pointer-events-none absolute"
                    style={{
                        width: expandSize,
                        height: size,
                    }}
                >
                    <div className="flex p-2 shrink-0" style={iconBoxStyle}>
                        {icon}
                    </div>
                    <p className="text-xs font-semibold shrink-0 pointer-events-none">
                        {label}
                    </p>
                    <div
                        className="flex absolute bottom-[-5px] justify-around"
                        style={{
                            width: expandSize,
                        }}
                    >
                        {[...Array(Math.ceil(expandSize / size)).keys()].map(
                            (n) => {
                                return (
                                    <Text
                                        fw={n === index ? 800 : ''}
                                        c={
                                            n === selectedIndex
                                                ? '#4444ff'
                                                : 'white'
                                        }
                                        key={n}
                                    >
                                        _
                                    </Text>
                                )
                            }
                        )}
                    </div>
                </div>
            )}
        </div>
    )
}

export default WeblensButton
