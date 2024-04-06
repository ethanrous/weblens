import { Box, Loader, Text } from '@mantine/core';
import { CSSProperties, memo, useMemo, useState } from 'react';
import { ColumnBox, RowBox } from '../Pages/FileBrowser/FilebrowserStyles';
import { IconCheck, IconX } from '@tabler/icons-react';
import { useResize } from './hooks';

type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | boolean | Promise<boolean>;

type buttonProps = {
    label: string;
    postScript?: string;
    showSuccess?: boolean;
    toggleOn?: boolean;
    subtle?: boolean;
    allowRepeat?: boolean;
    centerContent?: boolean;
    danger?: boolean;
    disabled?: boolean;
    Left?;
    Right?;

    // Style
    width?: string | number;
    height?: string | number;
    fontSize?: string;

    onClick?: ButtonActionHandler;
    onMouseUp?: ButtonActionHandler;
    onMouseOver?: ButtonActionHandler;
    style?: CSSProperties;
    setButtonRef?;
};

const ButtonContent = ({
    label,
    postScript,
    Left,
    Right,
    fontSize,
    showText,
    centerContent,
    setTextRef,
}) => {
    let leftWidth;
    if (Left && showText) {
        leftWidth = 34;
    } else if (showText) {
        leftWidth = 0;
    } else if (Left) {
        leftWidth = 24;
    }

    return (
        <RowBox
            style={{
                justifyContent:
                    showText && !centerContent ? 'flex-start' : 'center',
            }}
        >
            <Box
                style={{
                    height: 24,
                    width: leftWidth,
                    paddingRight: showText && Left ? 4 : 0,
                    paddingLeft: showText && Left ? 4 : 0,
                }}
            >
                {Left}
            </Box>
            <ColumnBox
                style={{
                    width: '100%',
                    height: 'max-content',
                    alignItems:
                        showText && !centerContent ? 'flex-start' : 'center',
                    justifyContent: 'center',
                }}
            >
                <Text
                    ref={setTextRef}
                    truncate="end"
                    fw={'inherit'}
                    size="inherit"
                    style={{
                        width: '100%',
                        padding: 2,
                        userSelect: 'none',
                        textWrap: 'nowrap',
                        flexGrow: 1,
                        fontSize: fontSize,
                        display: showText ? 'block' : 'none',
                    }}
                >
                    {label}
                </Text>

                {postScript && (
                    <Text
                        fw={300}
                        size="10px"
                        truncate="end"
                        style={{ padding: 2, textWrap: 'nowrap' }}
                    >
                        {postScript}
                    </Text>
                )}
            </ColumnBox>
            <Box
                style={{
                    height: 24,
                    width: leftWidth,
                    paddingRight: showText && Right ? 4 : 0,
                    paddingLeft: showText && Right ? 4 : 0,
                }}
            >
                {Right}
            </Box>
        </RowBox>
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
    try {
        setLoading(true);
        const res = await handler(e);
        setLoading(false);
        if (res === true && showSuccess) {
            setSuccess(true);
            setTimeout(() => setSuccess(false), 2000);
        } else if (res === false && showSuccess) {
            setFail(true);
            setTimeout(() => setFail(false), 2000);
        }
    } catch (e) {
        setLoading(false);
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
        allowRepeat = true,
        centerContent = false,
        disabled = false,
        danger = false,
        Left = null,
        Right = null,
        onClick,

        width,
        height,
        fontSize = '16px',

        onMouseUp,
        onMouseOver,
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

        const [textRef, setTextRef]: [
            textRef: HTMLDivElement,
            setTextRef: any
        ] = useState(null);
        const [textSize, setTextSize] = useState({ height: null, width: null });

        const showText = useMemo(() => {
            if ((!Left && !Right) || buttonSize.width == null) {
                return true;
            }

            return buttonSize.width > textSize.width + 60;
        }, [Left, Right, buttonSize.width, textSize.width]);

        return (
            <Box
                ref={(r) => {
                    setButtonRef(r);
                    setSizeRef(r);
                }}
                className={
                    toggleOn === undefined
                        ? 'weblens-button'
                        : 'weblens-toggle-button'
                }
                mod={{
                    'data-toggled':
                        toggleOn !== undefined ? toggleOn.toString() : '',
                    'data-repeat': allowRepeat.toString(),
                    'data-success': success.toString(),
                    'data-fail': fail.toString(),
                    'data-subtle': subtle.toString(),
                    'data-disabled': disabled.toString(),
                    'data-danger': danger.toString(),
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
                        setTextRef={setTextRef}
                    />
                )}
                {success && showSuccess && (
                    <ColumnBox style={{ width: 'max-content', height: '28px' }}>
                        <IconCheck />
                    </ColumnBox>
                )}
                {fail && showSuccess && (
                    <ColumnBox style={{ width: 'max-content', height: '28px' }}>
                        <IconX />
                    </ColumnBox>
                )}
                {loading && showSuccess && (
                    <ColumnBox
                        style={{
                            width: 'max-content',
                            height: '28px',
                            justifyContent: 'center',
                        }}
                    >
                        <Loader size={'20px'} color={'white'} />
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
