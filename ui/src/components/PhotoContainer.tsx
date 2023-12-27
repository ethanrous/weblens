import { useState, useEffect, useRef, useContext, useMemo } from "react";
import { Blurhash } from "react-blurhash";
import { userContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react"
import { AspectRatio, Box, CSSProperties, Image, Loader, MantineStyleProp } from "@mantine/core";


import API_ENDPOINT from '../api/ApiEndpoint'
import { itemData } from "../types/Types";

// Styles

const ThumbnailContainer = ({ reff, style, children }) => {
    return (
        <Box
            ref={reff}
            draggable={false}
            style={{
                height: '100%',
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                ...style,
            }}
            onDrag={(e) => { e.preventDefault(); e.stopPropagation() }}
            children={children}
        />
    )
}

//Components

export function useIsVisible(root, ref, maintained: boolean = false, margin: number = 100, thresh: number = 0.0) {
    const [isVisible, setIsVisible] = useState(false)
    const visibleRef = useRef(false)
    useEffect(() => {
        if (!ref.current) {
            return
        }
        let options: IntersectionObserverInit = {
            root: root?.current || null,
            rootMargin: `${margin}px`,
            threshold: thresh
        }

        const observer = new IntersectionObserver(([entry]) => {
            if (maintained && entry.isIntersecting) {
                visibleRef.current = entry.isIntersecting
                setIsVisible(entry.isIntersecting)
            } else if (!maintained) {
                visibleRef.current = entry.isIntersecting
                setIsVisible(entry.isIntersecting)
            }

        }, options)

        observer.observe(ref.current);
        return () => {
            observer.disconnect();
        };
    }, [ref.current]);

    return { isVisible: isVisible, visibleStateRef: visibleRef };
}

function getImageData(url, filehash, authHeader, signal) {
    return fetch(url, { headers: authHeader, signal }).then(res => res.blob()).then((blob) => {
        if (blob.length === 0) {
            Promise.reject("Empty blob")
        }
        return { data: URL.createObjectURL(blob), hash: filehash }
    }).catch((r) => {
        if (!signal.aborted)
        console.error("Failed to get image from server: ", r)
    })
}

function getImageMeta(url, filehash, authHeader, signal) {
    return fetch(url, { headers: authHeader, signal }).then(res => res.json()).then((json) => {
        return { data: json, hash: filehash }
    }).catch((r) => {
        if (!signal.aborted)
            console.error("Failed to get image from server: ", r)
    })
}

export const MediaImage = ({
    mediaId,
    quality,
    blurhash,
    lazy = true,
    expectFailure = false,
    containerStyle,
    imgStyle,
    root
}: { mediaId: string, quality: "thumbnail" | "fullres", blurhash?: string, lazy?: boolean, expectFailure?: boolean, containerStyle?: CSSProperties, imgStyle?: MantineStyleProp, root?}
) => {
    const [loaded, setLoaded] = useState(false)
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [imgData, setImgData] = useState(null)
    const [imgMeta, setImgMeta]: [imgMeta: itemData, setImgMeta: any] = useState(null)

    const [metaPromise, setMetaPromise] = useState(null)
    const [thumbPromise, setThumbPromise] = useState(null)
    const [fullresPromise, setFullresPromise] = useState(null)

    const visibleRef = useRef()
    const { isVisible, visibleStateRef } = useIsVisible(root, visibleRef, false, 1000, 0)

    const hashRef = useRef("")

    const abortController = new AbortController();
    const signal = abortController.signal;

    const metaUrl = new URL(`${API_ENDPOINT}/item/${mediaId}`)
    metaUrl.searchParams.append("meta", "true")
    const thumbUrl = new URL(`${API_ENDPOINT}/item/${mediaId}`)
    thumbUrl.searchParams.append("thumbnail", "true")
    const fullresUrl = new URL(`${API_ENDPOINT}/item/${mediaId}`)
    fullresUrl.searchParams.append("fullres", "true")

    useEffect(() => {
        setImgData(null)
        setMetaPromise(null)
        setThumbPromise(null)
        setFullresPromise(null)
        setLoaded(false)
        hashRef.current = mediaId
    }, [mediaId])

    metaPromise?.then((res) => {
        if (res && res.data && res.hash === hashRef.current && !loaded && !imgData && !loadError) {
            setImgMeta(res.data)
        } else if (res === undefined) {
            setMetaPromise(null)
        }
    })
    thumbPromise?.then((res) => {
        if (res && res.data && res.hash === hashRef.current && !loaded && !imgData && !loadError) {
            setImgData(res.data)
            if (quality === "thumbnail") {
                setLoaded(true)
            }
        } else if (res === undefined) {
            setThumbPromise(null)
        }
    })
    fullresPromise?.then((res) => {
        if (res && res.data && res.hash === hashRef.current && !loaded && !loadError) {
            setImgData(res.data)
            setLoaded(true)
        } else if (res === undefined) {
            setFullresPromise(null)
        }
    })

    useEffect(() => {
        if (hashRef.current === "") {
            setLoadErr(true)
        } else if (isVisible && !thumbPromise && !fullresPromise) {
            setThumbPromise(getImageData(thumbUrl, mediaId, authHeader, signal))
            if (quality === "fullres") {
                setFullresPromise(getImageData(fullresUrl, mediaId, authHeader, signal))
            }
        }
        return () => {
            if (!visibleStateRef.current || mediaId !== hashRef.current) {
                abortController.abort()
            }
        }
    }, [isVisible, hashRef.current])

    const sizer = useMemo(() => {
        if (!imgMeta) {
            return null
        }
        const sizer = imgMeta?.mediaData.mediaHeight > imgMeta?.mediaData.mediaWidth ? { width: '100%' } : { height: '100%' }
        return sizer
    }, [imgMeta])

    return (
        <ThumbnailContainer reff={visibleRef} style={containerStyle}>
            {(isVisible && loadError && !expectFailure) && (
                <IconExclamationCircle color="red" style={{ position: 'absolute' }} />
            )}
            {(isVisible && loadError && expectFailure) && (
                <IconPhoto style={{ position: 'absolute' }} />
            )}
            {(!lazy && isVisible && !loaded && !loadError) && (
                <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
            )}

            <Image
                draggable={false}
                src={imgData}
                style={{ ...sizer, display: imgData ? "block" : "none", userSelect: 'none', ...imgStyle }}
            />

            {isVisible && blurhash && lazy && !imgData && (
                <Blurhash
                    style={{ position: "absolute" }}
                    height={containerStyle?.height ? containerStyle?.height : 250}
                    width={550}
                    hash={blurhash}
                />
            )}

        </ThumbnailContainer >
    )
}