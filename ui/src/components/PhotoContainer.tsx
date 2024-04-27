import {
    useState,
    useEffect,
    useRef,
    useContext,
    memo,
    useCallback,
} from "react";
// import { Blurhash } from "react-blurhash";
import { MediaTypeContext, UserContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react";
import { Box, CSSProperties, Loader } from "@mantine/core";
import API_ENDPOINT, { PUBLIC_ENDPOINT } from "../api/ApiEndpoint";
import { AuthHeaderT, UserContextT } from "../types/Types";
import WeblensMedia from "../classes/Media";

import "./style.css";

function getImageData(
    url: string,
    mediaId: string,
    authHeader: AuthHeaderT,
    signal,
    setLoadErr
) {
    const res = fetch(url, { headers: authHeader, signal })
        .then((res) => {
            if (res.status !== 200) {
                return Promise.reject(res.statusText);
            }
            return res.arrayBuffer();
        })
        .then((buf) => {
            if (buf.byteLength === 0) {
                Promise.reject("Empty blob");
            }

            return { data: buf, hash: mediaId };
        })
        .catch((r) => {
            if (!signal.aborted) {
                console.error("Failed to get image from server:", r);
                setLoadErr(true);
            }
        });

    return res;
}

export const MediaImage = memo(
    ({
        media,
        setMediaCallback,
        quality,
        pageNumber = undefined,
        lazy = true,
        expectFailure = false,
        preventClick = false,
        doFetch = true,
        imgStyle,
        containerStyle,
        doPublic = false,
        disabled = false,
    }: {
        media: WeblensMedia;
        setMediaCallback?: (
            mediaId: string,
            quality: "thumbnail" | "fullres",
            data: ArrayBuffer
        ) => void;
        quality: "thumbnail" | "fullres";
        pageNumber?: number;
        lazy?: boolean;
        expectFailure?: boolean;
        preventClick?: boolean;
        doFetch?: boolean;
        imgStyle?: CSSProperties;
        containerStyle?: CSSProperties;
        doPublic?: boolean;
        disabled?: boolean;
    }) => {
        const [loadError, setLoadErr] = useState(false);
        const { authHeader }: UserContextT = useContext(UserContext);
        const typeMap = useContext(MediaTypeContext);
        const [imgData, setImgData] = useState(null);
        const visibleRef = useRef(null);
        const hashRef = useRef("");
        const abortController = new AbortController();

        if (!media) {
            media = new WeblensMedia({ mediaId: "" });
        }

        console.log(media);

        const fetchFullres = useCallback(async () => {
            const url = new URL(
                `${
                    doPublic ? PUBLIC_ENDPOINT : API_ENDPOINT
                }/media/${media.Id()}/fullres`
            );
            if (pageNumber !== undefined) {
                url.searchParams.append("page", pageNumber.toString());
            }
            const ret = await getImageData(
                url.toString(),
                media.Id(),
                authHeader,
                abortController.signal,
                setLoadErr
            );
            if (!ret) {
                return;
            }
            if (setMediaCallback) {
                setMediaCallback(media.Id(), "fullres", ret.data);
            } else if (pageNumber === undefined) {
                media.SetFullresBytes(ret.data);
            }
            setImgData(URL.createObjectURL(new Blob([ret.data])));
        }, [media.Id(), authHeader]);

        const fetchThumbnail = useCallback(async () => {
            if (pageNumber !== undefined && pageNumber > 0) {
                return;
            }
            const ret = await getImageData(
                `${
                    doPublic ? PUBLIC_ENDPOINT : API_ENDPOINT
                }/media/${media.Id()}/thumbnail`,
                media.Id(),
                authHeader,
                abortController.signal,
                setLoadErr
            );
            if (!ret) {
                return;
            }
            if (setMediaCallback) {
                setMediaCallback(media.Id(), "thumbnail", ret.data);
            } else if (pageNumber === undefined) {
                media.SetThumbnailBytes(ret.data);
            }
            setImgData((prev) =>
                prev === "" ? URL.createObjectURL(new Blob([ret.data])) : prev
            );
            // setLoaded((prev) => (prev === "" ? "thumbnail" : prev));
        }, [media.Id(), authHeader]);

        useEffect(() => {
            setLoadErr(false);
            if (!media.Id()) {
                return;
            }
            hashRef.current = media.Id();

            if (
                media.HighestQualityLoaded() == "fullres" &&
                quality === "fullres"
            ) {
                setImgData(media.GetImgUrl("fullres"));
                // setLoaded("fullres");
            } else if (
                media.HighestQualityLoaded() == "thumbnail" &&
                (pageNumber === undefined || pageNumber === 0)
            ) {
                setImgData(media.GetImgUrl("thumbnail"));
                // setLoaded("thumbnail");
            } else {
                setImgData("");
            }

            if (
                media.HighestQualityLoaded() !== "fullres" &&
                doFetch &&
                quality === "fullres"
            ) {
                fetchFullres();
            }

            if (!media.HighestQualityLoaded() && doFetch) {
                fetchThumbnail();
            }

            return () => abortController.abort();
        }, [media.Id()]);

        return (
            <Box
                className="photo-container"
                ref={visibleRef}
                style={{ ...containerStyle }}
                onDrag={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                }}
                onClick={(e) => {
                    preventClick && e.stopPropagation();
                }}
            >
                {loadError && !expectFailure && (
                    <IconExclamationCircle color="red" />
                )}
                {((loadError && expectFailure) || !media.Id()) && <IconPhoto />}
                {quality === "fullres" &&
                    media.HighestQualityLoaded() !== "fullres" &&
                    !loadError && (
                        <Loader
                            color="white"
                            bottom={40}
                            right={40}
                            size={20}
                            style={{ position: "absolute" }}
                        />
                    )}

                <img
                    alt=""
                    className={
                        quality === "thumbnail"
                            ? "media-thumbnail"
                            : "media-fullres"
                    }
                    draggable={false}
                    src={imgData}
                    style={{
                        display: imgData && !loadError ? "" : "none",
                        filter: disabled ? "grayscale(100%)" : "",
                        zIndex: "inherit",
                        position: "relative",
                        ...imgStyle,
                    }}
                />

                {quality === "fullres" && media.GetMediaType().IsVideo && (
                    <video src="" controls />
                )}

                {/* {media?.blurHash && lazy && !imgData && (
                    <Blurhash
                        style={{ position: "absolute" }}
                        height={visibleRef?.current?.clientHeight ? visibleRef.current.clientHeight : 0}
                        width={visibleRef?.current?.clientWidth ? visibleRef.current.clientWidth : 0}
                        hash={media.blurHash}
                    />
                )} */}
            </Box>
        );
    },
    (last, next) => {
        if (last.media?.Id() !== next.media?.Id()) {
            return false;
        } else if (last.containerStyle !== next.containerStyle) {
            return false;
        } else if (
            last.media?.HighestQualityLoaded() !==
            next.media?.HighestQualityLoaded()
        ) {
            return false;
        }
        return true;
    }
);
