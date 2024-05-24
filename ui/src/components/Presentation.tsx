import { memo, useCallback, useEffect, useMemo, useState } from "react";

import { MediaImage } from "./PhotoContainer";
import { Box } from "@mantine/core";
import { useMediaType } from "./hooks";
import WeblensMedia from "../classes/Media";
import { IconX } from "@tabler/icons-react";
import { PresentType, SizeT } from "../types/Types";

export const PresentationContainer = ({
    shadeOpacity,
    onMouseMove,
    onClick,
    children,
}: {
    shadeOpacity?;
    onMouseMove?;
    onClick?;
    children;
}) => {
    if (!shadeOpacity) {
        shadeOpacity = "0.90";
    }
    return (
        <Box
            onMouseMove={onMouseMove}
            onClick={onClick}
            style={{
                position: "fixed",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                top: 0,
                left: 0,
                padding: "25px",
                height: "100%",
                width: "100%",
                zIndex: 100,
                backgroundColor: `rgb(0, 0, 0, ${shadeOpacity})`,
                backdropFilter: "blur(4px)",
            }}
            children={children}
        />
    );
};

export function GetMediaFullscreenSize(
    mediaData: WeblensMedia,
    containerSize: SizeT
): SizeT {
    let newWidth;
    if (!containerSize) {
        newWidth = 0;
    } else if (containerSize.width < 150 && mediaData.GetPageCount() > 1) {
        newWidth = 150;
    } else {
        newWidth = containerSize.width;
    }

    if (
        !mediaData ||
        !mediaData.GetHeight() ||
        !mediaData.GetWidth() ||
        !containerSize.height ||
        !containerSize.width
    ) {
        return { height: 0, width: 0 };
    }
    const mediaRatio = mediaData.GetWidth() / mediaData.GetHeight();
    const windowRatio = containerSize.width / containerSize.height;
    let absHeight = 0;
    let absWidth = 0;
    if (mediaRatio > windowRatio) {
        absWidth = containerSize.width;
        absHeight = (absWidth / mediaData.GetWidth()) * mediaData.GetHeight();
    } else {
        absHeight = containerSize.height;
        absWidth = (absHeight / mediaData.GetHeight()) * mediaData.GetWidth();
    }
    return { height: absHeight, width: absWidth };
}

export const ContainerMedia = ({
    mediaData,
    containerRef,
}: {
    mediaData: WeblensMedia;
    containerRef;
}) => {
    const [boxSize, setBoxSize] = useState({
        height: containerRef?.clientHeight || 0,
        width: containerRef?.clientWidth || 0,
    });
    const [, forceUpdate] = useState(false);
    useEffect(() => {
        if (containerRef) {
            const obs = new ResizeObserver((entries) => {
                forceUpdate((p) => !p);
            });
            obs.observe(containerRef);
            return () => obs.disconnect();
        }
    }, [containerRef]);

    useEffect(() => {
        let newWidth;
        if (!containerRef) {
            newWidth = 0;
        } else if (
            containerRef.clientWidth < 150 &&
            mediaData.GetPageCount() > 1
        ) {
            newWidth = 150;
        } else {
            newWidth = containerRef.clientWidth;
        }
        setBoxSize({ height: containerRef?.clientHeight, width: newWidth });
    }, [containerRef?.clientWidth, containerRef?.clientHeight]);

    const [absHeight, absWidth] = useMemo(() => {
        if (
            !mediaData ||
            !mediaData.GetHeight() ||
            !mediaData.GetWidth() ||
            !boxSize.height ||
            !boxSize.width
        ) {
            return [0, 0];
        }
        const mediaRatio = mediaData.GetWidth() / mediaData.GetHeight();
        const windowRatio = boxSize.width / boxSize.height;
        let absHeight = 0;
        let absWidth = 0;
        if (mediaRatio > windowRatio) {
            absWidth = boxSize.width;
            absHeight =
                (absWidth / mediaData.GetWidth()) * mediaData.GetHeight();
        } else {
            absHeight = boxSize.height;
            absWidth =
                (absHeight / mediaData.GetHeight()) * mediaData.GetWidth();
        }
        return [absHeight, absWidth];
    }, [mediaData, mediaData.GetHeight(), mediaData.GetWidth(), boxSize]);

    if (!mediaData || !containerRef) {
        return null;
    }

    if (mediaData.GetPageCount() > 1) {
        return (
            <Box
                className="no-scrollbar"
                style={{ overflow: "scroll", gap: absHeight * 0.02 }}
            >
                {[...Array(mediaData.GetPageCount()).keys()].map((p) => (
                    <MediaImage
                        key={p}
                        media={mediaData}
                        quality={"fullres"}
                        pageNumber={p}
                        containerStyle={{ height: absHeight, width: absWidth }}
                        preventClick
                    />
                ))}
            </Box>
        );
    } else {
        return (
            <MediaImage
                media={mediaData}
                quality={"fullres"}
                containerStyle={{ height: absHeight, width: absWidth }}
                preventClick
            />
        );
    }
};

const PresentationVisual = ({
    mediaData,
    Element,
}: {
    mediaData: WeblensMedia;
    Element;
}) => {
    const [containerRef, setContainerRef] = useState(null);
    return (
        <Box
            style={{
                height: "100%",
                width: "100%",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-around",
            }}
        >
            {mediaData && (
                <Box
                    style={{
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        width: Element ? "50%" : "100%",
                        height: "100%",
                    }}
                    ref={setContainerRef}
                >
                    <ContainerMedia
                        mediaData={mediaData}
                        containerRef={containerRef}
                    />
                </Box>
            )}
            {Element && <Element />}
        </Box>
    );
};

function useKeyDownPresentation(itemId: string, dispatch) {
    const keyDownHandler = useCallback(
        (event) => {
            if (!itemId) {
                return;
            } else if (event.key === "Escape") {
                event.preventDefault();
                dispatch({ type: "stop_presenting" });
            } else if (event.key === "ArrowLeft") {
                event.preventDefault();
                dispatch({ type: "presentation_previous" });
            } else if (event.key === "ArrowRight") {
                event.preventDefault();
                dispatch({ type: "presentation_next" });
            } else if (event.key === "ArrowUp" || event.key === "ArrowDown") {
                event.preventDefault();
            }
        },
        [itemId, dispatch]
    );
    useEffect(() => {
        window.addEventListener("keydown", keyDownHandler);
        return () => {
            window.removeEventListener("keydown", keyDownHandler);
        };
    }, [keyDownHandler]);
}

function handleTimeout(to, setTo, setGuiShown) {
    if (to) {
        clearTimeout(to);
    }
    setTo(setTimeout(() => setGuiShown(false), 1000));
}

const Presentation = memo(
    ({
        itemId,
        mediaData,
        element,
        dispatch,
    }: {
        itemId: string;
        mediaData: WeblensMedia;
        dispatch: React.Dispatch<any>;
        element?;
    }) => {
        useKeyDownPresentation(itemId, dispatch);

        const [to, setTo] = useState(null);
        const [guiShown, setGuiShown] = useState(false);

        if (!mediaData) {
            return null;
        }

        return (
            <PresentationContainer
                onMouseMove={() => {
                    setGuiShown(true);
                    handleTimeout(to, setTo, setGuiShown);
                }}
                onClick={() =>
                    dispatch({ type: "set_presentation", media: null })
                }
            >
                <PresentationVisual mediaData={mediaData} Element={element} />
                {/* <Text style={{ position: 'absolute', bottom: guiShown ? 15 : -100, left: '50vw' }} >{}</Text> */}

                <Box
                    className="close-icon"
                    mod={{ shown: guiShown.toString() }}
                    onClick={() =>
                        dispatch({
                            type: "set_presentation",
                            presentingId: null,
                        })
                    }
                >
                    <IconX />
                </Box>
            </PresentationContainer>
        );
    },
    (prev, next) => {
        if (prev.itemId !== next.itemId) {
            return false;
        }

        return true;
    }
);

export default Presentation;
