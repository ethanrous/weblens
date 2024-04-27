import { Box, Loader, Text } from "@mantine/core";
import { CSSProperties, memo, useEffect, useMemo, useState } from "react";
import { ColumnBox } from "../Pages/FileBrowser/FileBrowserStyles";
import { IconCheck, IconX } from "@tabler/icons-react";
import { useResize } from "./hooks";

import "./weblensButton.scss";

type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | boolean | Promise<boolean>;

type buttonProps = {
    label?: string;
    postScript?: string;
    showSuccess?: boolean;
    toggleOn?: boolean;
    subtle?: boolean;
    allowRepeat?: boolean;
    centerContent?: boolean;
    danger?: boolean;
    disabled?: boolean;
    doSuper?: boolean;
    Left?: JSX.Element;
    Right?: JSX.Element;

    // Style
    height: number;
    width?: number | string;
    fontSize?: number;

    onClick?: ButtonActionHandler;
    onMouseUp?: ButtonActionHandler;
    onMouseOver?: ButtonActionHandler;
    onMouseLeave?: ButtonActionHandler;
    style?: CSSProperties;
    setButtonRef?;
};

const ButtonContent = ({
    label,
    postScript,
    Left,
    Right,
    buttonWidth,
    fontSize,
    iconSize,
    centerContent,
}: {
    label: string;
    postScript: string;
    Left: JSX.Element;
    Right: JSX.Element;
    buttonWidth: number;
    fontSize: number;
    iconSize: number;
    centerContent: boolean;
}) => {
    const [showText, setShowText] = useState(Boolean(label));

    useEffect(() => {
        if (
            !Boolean(label) ||
            ((Boolean(Left) || Boolean(Right)) &&
                buttonWidth < iconSize + label?.length &&
                buttonWidth !== 0)
        ) {
            setShowText(false);
        } else {
            setShowText(true);
        }
    }, [buttonWidth]);

    if (!iconSize) {
        iconSize = 24;
    }

    return (
        <Box
            style={{
                height: "100%",
                display: "flex",
                flexDirection: "row",
                justifyContent: centerContent ? "center" : "flex-start",
                alignItems: "center",
                width: "100%",
                minWidth: 24,
            }}
        >
            <Box
                style={{
                    flexShrink: 0,
                    maxHeight: "100%",
                    height: iconSize,
                    width: iconSize,
                    marginRight: showText && Left ? 4 : 0,
                    marinLeft: showText && Left ? 4 : 0,
                    display: Left ? "flex" : "none",
                    alignItems: "center",
                }}
            >
                {Left}
            </Box>
            <ColumnBox
                style={{
                    width: "max-content",
                    height: "100%",
                    alignItems: centerContent ? "center" : "flex-start",
                    justifyContent: "center",
                }}
            >
                <p
                    // ref={setTextRef}
                    // truncate="end"
                    style={{
                        margin: 0,
                        flexGrow: 0,
                        flexShrink: 1,
                        padding: "2px",
                        fontSize: fontSize,
                        userSelect: "none",
                        textWrap: "nowrap",
                        lineHeight: "10px",
                        width: "max-content",
                        fontWeight: "inherit",
                        height: "max-content",
                        display: showText ? "block" : "none",
                    }}
                >
                    {label}
                </p>

                {postScript && (
                    <Text
                        fw={300}
                        size="10px"
                        truncate="end"
                        style={{
                            padding: 2,
                            textWrap: "nowrap",
                            overflow: "visible",
                            userSelect: "none",
                        }}
                    >
                        {postScript}
                    </Text>
                )}
            </ColumnBox>
            {Boolean(Right) && (
                <Box
                    style={{
                        flexShrink: 0,
                        maxHeight: "100%",
                        height: iconSize,
                        width: iconSize,
                        marginRight: showText && Right ? 4 : 0,
                        marginLeft: showText && Right ? 4 : 0,
                        display: Right ? "flex" : "none",
                        alignItems: "center",
                    }}
                >
                    {Right}
                </Box>
            )}
        </Box>
    );
};

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
        return;
    }
    const tm = setTimeout(() => {
        setLoading(true);
    }, 150);
    try {
        // Don't flash loading if handler returns instantly
        const res = await handler(e);
        clearTimeout(tm);
        setLoading(false);
        if (res === true && showSuccess) {
            setSuccess(true);
            setTimeout(() => setSuccess(false), 2000);
        } else if (res === false && showSuccess) {
            setFail(true);
            setTimeout(() => setFail(false), 2000);
        }
    } catch (e) {
        clearTimeout(tm);
        setLoading(false);
        console.error(e);
        if (showSuccess) {
            setSuccess(false);
            setFail(true);
            setTimeout(() => setFail(false), 2000);
        }
    }
};

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
        fontSize = 20,

        onMouseUp,
        onMouseOver,
        onMouseLeave,
        style,
        setButtonRef = (r) => {},
    }: buttonProps) => {
        const [success, setSuccess] = useState(false);
        const [fail, setFail] = useState(false);
        const [loading, setLoading] = useState(false);

        const [sizeRef, setSizeRef]: [
            buttonRef: HTMLDivElement,
            setButtonRef: any
        ] = useState(null);
        const buttonSize = useResize(sizeRef);

        const [contentRef, setContentRef]: [
            contentRef: HTMLDivElement,
            setContentRef: any
        ] = useState(null);
        const contentSize = useResize(contentRef);

        const mainWidth = useMemo(() => {
            if (success) {
                return height;
            } else if (width) {
                return width;
            } else {
                return "max-content";
            }
        }, [contentSize.width, height, width]);

        return (
            <Box
                key={label}
                ref={(r) => {
                    setButtonRef(r);
                    setSizeRef(r);
                }}
                className={
                    toggleOn === undefined
                        ? "weblens-button"
                        : "weblens-toggle-button"
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
                    width: mainWidth,
                    height: height,
                    minWidth: height,
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
                <Box ref={setContentRef} style={{ display: "flex" }}>
                    {!success && !fail && !loading && (
                        <ButtonContent
                            label={label}
                            postScript={postScript}
                            Left={Left}
                            Right={Right}
                            buttonWidth={buttonSize.width}
                            fontSize={fontSize ? fontSize : height * 0.5}
                            iconSize={height * 0.6}
                            centerContent={centerContent}
                        />
                    )}
                    {success && showSuccess && <IconCheck />}
                    {fail && showSuccess && (
                        <ColumnBox
                            style={{
                                width: "max-content",
                                height: "max-content",
                            }}
                        >
                            <IconX />
                        </ColumnBox>
                    )}
                    {loading && showSuccess && (
                        <ColumnBox
                            style={{
                                width: "max-content",
                                height: "max-content",
                                justifyContent: "center",
                            }}
                        >
                            <Loader size={"20px"} color={"white"} />
                        </ColumnBox>
                    )}
                </Box>
            </Box>
        );
    },
    (prev, next) => {
        if (prev.toggleOn !== next.toggleOn) {
            return false;
        } else if (prev.label !== next.label) {
            return false;
        } else if (prev.disabled !== next.disabled) {
            return false;
        } else if (prev.onClick !== next.onClick) {
            return false;
        } else if (prev.onMouseUp !== next.onMouseUp) {
            return false;
        } else if (prev.onMouseOver !== next.onMouseOver) {
            return false;
        } else if (prev.postScript !== next.postScript) {
            return false;
        } else if (prev.width !== next.width) {
            return false;
        } else if (prev.height !== next.height) {
            return false;
        } else if (prev.style !== next.style) {
            return false;
        }
        return true;
    }
);

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
    size: number;
    expandSize?: number;
    label?: string;
    icon: JSX.Element;
    selected: boolean;
    index?: number;
    selectedIndex?: number;
    onClick?;
}) => {
    const [hover, setHover] = useState(false);
    return (
        <Box>
            <Box
                className="weblens-select-icon"
                mod={{ "data-selected": selected }}
                style={{ height: size, width: size }}
                onClick={(e) => {
                    onClick(e);
                }}
                onMouseOver={(e) => setHover(true)}
                onMouseLeave={(e) => setHover(false)}
            >
                {icon}
            </Box>
            {hover && expandSize && (
                <Box
                    style={{
                        pointerEvents: "none",
                        position: "absolute",
                        display: "flex",
                        alignItems: "center",

                        left: 0,
                        top: 0,
                        zIndex: 1,
                        width: expandSize,
                        height: size,
                        backgroundColor: "#222222",
                    }}
                >
                    <Box
                        style={{
                            height: size,
                            width: size,
                            display: "flex",
                            padding: 6,
                            flexShrink: 0,
                        }}
                    >
                        {icon}
                    </Box>
                    <Text
                        size="11px"
                        fw={600}
                        style={{ flexShrink: 0, pointerEvents: "none" }}
                    >
                        {label}
                    </Text>
                    <Box
                        style={{
                            position: "absolute",
                            bottom: -5,
                            display: "flex",
                            width: expandSize,
                            justifyContent: "space-around",
                        }}
                    >
                        {[...Array(Math.ceil(expandSize / size)).keys()].map(
                            (n) => {
                                return (
                                    <Text
                                        fw={n === index ? 800 : ""}
                                        c={
                                            n === selectedIndex
                                                ? "#4444ff"
                                                : "white"
                                        }
                                        key={n}
                                    >
                                        _
                                    </Text>
                                );
                            }
                        )}
                    </Box>
                </Box>
            )}
        </Box>
    );
};
