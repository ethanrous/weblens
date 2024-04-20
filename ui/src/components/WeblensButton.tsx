import { Box, Loader, Text } from "@mantine/core";
import { CSSProperties, memo, useEffect, useMemo, useState } from "react";
import { ColumnBox } from "../Pages/FileBrowser/FileBrowserStyles";
import { IconCheck, IconX } from "@tabler/icons-react";
import { useResize } from "./hooks";

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
    width?: string | number;
    height?: string | number;
    fontSize?: string;

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
    fontSize,
    // showText,
    centerContent,
}: {
    label: string;
    postScript: string;
    Left: JSX.Element;
    Right: JSX.Element;
    fontSize: string;
    showText: boolean;
    centerContent: boolean;
}) => {
    const [showText, setShowText] = useState(true);
    const [textRef, setTextRef] = useState(null);
    const textSize = useResize(textRef);
    useEffect(() => {
        if (textSize.width < 100) {
            setShowText(false);
        } else {
            setShowText(true);
        }
    }, [textSize.width]);

    return (
        <Box
            ref={setTextRef}
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
                    height: showText ? textSize.height : 24,
                    width: showText ? textSize.height : 24,
                    marginRight: showText && Left ? 4 : 0,
                    marinLeft: showText && Left ? 4 : 0,
                    display: Left ? "block" : "none",
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
                        fontWeight: "inherit",
                        width: "max-content",
                        height: "max-content",
                        flexShrink: 1,
                        padding: "2px",
                        userSelect: "none",
                        textWrap: "nowrap",
                        lineHeight: "10px",
                        flexGrow: 0,
                        fontSize: fontSize,
                        display: showText ? "block" : "none",
                        margin: 0,
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
                        height: textSize.height,
                        width: textSize.height,
                        marginRight: Right ? 4 : 0,
                        marginLeft: showText && Right ? 4 : 0,
                        display: Right ? "block" : "none",
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
        fontSize = "16px",

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

        const showText = useMemo(() => {
            if ((!Left && !Right) || buttonSize.width == null) {
                return true;
            }
            if (!label) {
                return false;
            }

            return buttonSize.width > 24;
        }, [Left, Right, buttonSize.width]);

        return (
            <Box
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
                    "data-toggled":
                        toggleOn !== undefined ? toggleOn.toString() : "",
                    "data-repeat": allowRepeat.toString(),
                    "data-success": success.toString(),
                    "data-fail": fail.toString(),
                    "data-subtle": subtle.toString(),
                    "data-disabled": disabled.toString(),
                    "data-super": doSuper.toString(),
                    "data-danger": danger.toString(),
                }}
                style={{ width: width, height: height, ...style }}
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
                {!success && !fail && !loading && (
                    <ButtonContent
                        label={label}
                        postScript={postScript}
                        Left={Left}
                        Right={Right}
                        fontSize={fontSize}
                        showText={showText}
                        centerContent={centerContent}
                    />
                )}
                {success && showSuccess && (
                    <ColumnBox
                        style={{ width: "max-content", height: "max-content" }}
                    >
                        <IconCheck />
                    </ColumnBox>
                )}
                {fail && showSuccess && (
                    <ColumnBox
                        style={{ width: "max-content", height: "max-content" }}
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
        );
        //     if (toggleOn === undefined) {
        //     } else {
        //         return (
        //             <Box
        //                 ref={(r) => {
        //                     setButtonRef(r);
        //                     setSizeRef(r);
        //                 }}
        //                 className="weblens-toggle-button"
        //                 mod={{
        //                     "data-toggled": toggleOn.toString(),
        //                     "data-repeat": allowRepeat.toString(),
        //                     "data-disabled": disabled.toString(),
        //                 }}
        //                 style={style}
        //                 onClick={() => onToggle(!toggleOn)}
        //             >
        //                 <Box
        //                     style={{
        //                         height: "24px",
        //                         paddingRight: showText ? 5 : 0,
        //                     }}
        //                 >
        //                     {Left}
        //                 </Box>
        //                 <ColumnBox
        //                     style={{
        //                         width: "100%",
        //                         alignItems: showText && !centerContent ? "flex-start" : "center",
        //                         paddingLeft: showText ? 10 : 0,
        //                     }}
        //                 >
        //                     <RowBox>
        //                         <Text
        //                             truncate="end"
        //                             ref={(e) => {
        //                                 if (!e || e.className === "weblens-toggle-button" || Boolean(textRef)) {
        //                                     return;
        //                                 }

        //                                 setTextRef(e);
        //                             }}
        //                             fw={"inherit"}
        //                             size="inherit"
        //                             style={{
        //                                 padding: 2,
        //                                 userSelect: "none",
        //                                 textWrap: "nowrap",
        //                                 width: 0,
        //                                 flexGrow: 1,
        //                                 display: showText ? "block" : "none",
        //                             }}
        //                         >
        //                             {label}
        //                         </Text>
        //                         {Right}
        //                     </RowBox>
        //                     {postScript && (
        //                         <Text fw={300} size="10px" style={{ padding: 2, userSelect: "none" }}>
        //                             {postScript}
        //                         </Text>
        //                     )}
        //                 </ColumnBox>
        //             </Box>
        //         );
        //     }
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
