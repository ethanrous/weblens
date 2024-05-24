import { Box, Loader, Text } from '@mantine/core'
import { CSSProperties, memo, useEffect, useMemo, useState } from 'react'
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
    Left?: JSX.Element
    Right?: JSX.Element

    // Style
    height: number
    width?: number | string
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
        buttonWidth,
        fontSize,
        iconSize,
        textFits,
        centerContent,
    }: {
        label: string
        postScript: string
        Left: JSX.Element
        Right: JSX.Element
        buttonWidth: number
        fontSize: string
        iconSize: number
        textFits: boolean
        centerContent: boolean
    }) => {
        const [showText, setShowText] = useState(Boolean(label) && textFits)

        useEffect(() => {
            if (
                !Boolean(label) ||
                ((Boolean(Left) || Boolean(Right)) &&
                    buttonWidth < iconSize + label?.length * 10 &&
                    buttonWidth !== 0)
            ) {
                if (showText) {
                    setShowText(false)
                }
            } else if (!showText && textFits) {
                setShowText(true)
            }
        }, [buttonWidth])

        if (!iconSize) {
            iconSize = 24
        }

        return (
            <div
                className={`flex flex-row h-full w-full items-center min-w-6 grow-0 justify-${
                    centerContent ? 'center' : 'flex-start'
                }`}
            >
                <div
                    className="shrink-0 max-h-max h-max w-max items-center justify-center"
                    style={{
                        height: iconSize,
                        width: iconSize,
                        marginRight: showText && Left ? 4 : 0,
                        marginLeft: showText && Left ? 4 : 0,
                        display: Left ? 'flex' : 'none',
                    }}
                >
                    {Left}
                </div>
                {(Boolean(label) || Boolean(postScript)) && (
                    <div
                        className="flex flex-col w-full h-full justify-center text-nowrap p-1"
                        style={{
                            alignItems: centerContent ? 'center' : 'flex-start',
                        }}
                    >
                        <p
                            className={`text-${fontSize} ${
                                showText ? 'block' : 'hidden'
                            } select-none w-max h-max shrink text-ellipsis whitespace-nowrap overflow-hidden`}
                        >
                            {label}
                        </p>

                        {postScript && (
                            <p className="font-light text-xs text-ellipsis text-nowrap overflow-visible select-none">
                                {postScript}
                            </p>
                        )}
                    </div>
                )}
                {Boolean(Right) && (
                    <div
                        className="shrink-0 max-h-max items-center"
                        style={{
                            height: iconSize,
                            width: iconSize,
                            marginRight: showText && Right ? 4 : 0,
                            marginLeft: showText && Right ? 4 : 0,
                            display: Right ? 'flex' : 'none',
                        }}
                    >
                        {Right}
                    </div>
                )}
            </div>
        )
    },
    (prev, next) => {
        if (prev.buttonWidth !== next.buttonWidth) {
            return false
        }
        if (prev.label !== next.label) {
            return false
        }
        if (prev.postScript !== next.postScript) {
            return false
        }
        if (prev.textFits !== next.textFits) {
            return false
        }
        if (prev.fontSize !== next.fontSize) {
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

export const WeblensButton = memo(
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
        Left = null,
        Right = null,
        onClick,

        width,
        height,
        fontSize = 'md',
        textMin,

        onMouseUp,
        onMouseOver,
        onMouseLeave,
        style,
        setButtonRef = (r) => {},
    }: buttonProps) => {
        const [success, setSuccess] = useState(false)
        const [fail, setFail] = useState(false)
        const [loading, setLoading] = useState(false)

        const [sizeRef, setSizeRef]: [
            buttonRef: HTMLDivElement,
            setButtonRef: any,
        ] = useState(null)
        const buttonSize = useResize(sizeRef)

        const [contentRef, setContentRef]: [
            contentRef: HTMLDivElement,
            setContentRef: any,
        ] = useState(null)
        const contentSize = useResize(contentRef)

        // const mainWidth = useMemo(() => {
        //     if (success) {
        //         return height;
        //     } else if (width) {
        //         return width;
        //     } else {
        //         return "max-content";
        //     }
        // }, [contentSize.width, height, width]);

        return (
            <div
                ref={setSizeRef}
                className="flex flex-col w-full items-center justify-center m-1 shrink grow"
                style={{ maxWidth: width, maxHeight: height }}
            >
                <Box
                    key={label}
                    ref={setButtonRef}
                    className={
                        toggleOn === undefined
                            ? 'weblens-button'
                            : 'weblens-toggle-button'
                    }
                    mod={{
                        toggled: (!!toggleOn).toString(),
                        repeat: allowRepeat.toString(),
                        success: success.toString(),
                        fail: fail.toString(),
                        subtle: subtle.toString(),
                        disabled: disabled.toString(),
                        super: doSuper.toString(),
                        danger: danger.toString(),
                    }}
                    style={{
                        height: height,
                        minWidth: height,
                        width: '100%',
                        ...style,
                    }}
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
                    <Box ref={setContentRef} className="flex w-full">
                        {!success && !fail && !loading && (
                            <ButtonContent
                                label={label}
                                postScript={postScript}
                                Left={Left}
                                Right={Right}
                                buttonWidth={buttonSize.width}
                                fontSize={fontSize}
                                textFits={
                                    textMin
                                        ? textMin <
                                          buttonSize.width - height * 0.6
                                        : true
                                }
                                iconSize={height * 0.6}
                                centerContent={centerContent}
                            />
                        )}
                        {success && showSuccess && <IconCheck />}
                        {fail && showSuccess && (
                            <div className="h-max w-max">
                                <IconX />
                            </div>
                        )}
                        {loading && showSuccess && (
                            <Box
                                style={{
                                    width: 'max-content',
                                    height: 'max-content',
                                    justifyContent: 'center',
                                }}
                            >
                                <Loader size={'20px'} color={'white'} />
                            </Box>
                        )}
                    </Box>
                </Box>
            </div>
        )
    },
    (prev, next) => {
        if (prev.toggleOn !== next.toggleOn) {
            console.log(next.label, 'HERE1')
            return false
        } else if (prev.label !== next.label) {
            console.log(next.label, 'HERE2')
            return false
        } else if (prev.disabled !== next.disabled) {
            console.log(next.label, 'HERE3')
            return false
        } else if (prev.onClick !== next.onClick) {
            // console.log(next.label, "HERE4");
            return false
        } else if (prev.onMouseUp !== next.onMouseUp) {
            console.log(next.label, 'HERE5')
            return false
        } else if (prev.onMouseOver !== next.onMouseOver) {
            console.log(next.label, 'HERE6')
            return false
        } else if (prev.postScript !== next.postScript) {
            console.log(next.label, 'HERE7')
            return false
        } else if (prev.width !== next.width) {
            console.log(next.label, 'HERE8')
            return false
        } else if (prev.height !== next.height) {
            console.log(next.label, 'HERE9')
            return false
        } else if (prev.style !== next.style) {
            console.log(next.label, 'HERE10')
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
