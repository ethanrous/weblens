import { useState, useEffect, useRef, useContext, useMemo, memo, useCallback } from "react";
import { Blurhash } from "react-blurhash";
import { userContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react"
import { Box, CSSProperties, Image, Loader, MantineStyleProp } from "@mantine/core";


import API_ENDPOINT from '../api/ApiEndpoint'
import { MediaData } from "../types/Types";

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
        if (!ref?.current) {
            return
        }

        const options: IntersectionObserverInit = {
            root: root?.current || null,

            rootMargin: "1000px 0px 1000px 0px",
            // rootMargin: `${margin}px`,
            // threshold: thresh
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
    }, [ref, root, maintained]);

    return { isVisible: isVisible, visibleStateRef: visibleRef };
}

function getImageData(url, filehash, authHeader, signal, setLoadErr) {
    const res = fetch(url, { headers: authHeader, signal })
    .then(res => {
        if (res.status !== 200) {
            return Promise.reject(res.statusText)
        }
        return res.blob()
    })
    .then(blob => {
        if (blob.length === 0) {
            Promise.reject("Empty blob")
        }
        return { data: URL.createObjectURL(blob), hash: filehash }
    })
    .catch((r) => {
        if (!signal.aborted) {
            console.error("Failed to get image from server:", r)
            setLoadErr(true)
        }
    })

    return res
}

function getImageMeta(url, filehash, authHeader, signal, setLoadErr) {
    return fetch(url, { headers: authHeader, signal }).then(res => res.json()).then((json) => {
        return { data: json, hash: filehash }
    }).catch((r) => {
        if (!signal.aborted) {
            console.error("Failed to get image meta from server: ", r)
            setLoadErr(true)
        }
    })
}

export const MediaImage = memo(({
    mediaId,
    quality,
    blurhash,
    metadataPreload,
    lazy = true,
    expectFailure = false,
    containerStyle,
    imgStyle,
    root
}: { mediaId: string, quality: "thumbnail" | "fullres", blurhash?: string, metadataPreload?: MediaData, lazy?: boolean, expectFailure?: boolean, containerStyle?: CSSProperties, imgStyle?: MantineStyleProp, root?}
) => {
    const [loaded, setLoaded] = useState("")
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [imgData, setImgData] = useState(null)
    const [imgMeta, setImgMeta]: [imgMeta: MediaData, setImgMeta: any] = useState(metadataPreload)

    const [metaPromise, setMetaPromise] = useState(null)
    const [thumbPromise, setThumbPromise] = useState(null)
    const [fullresPromise, setFullresPromise] = useState(null)

    const visibleRef = useRef()
    const { isVisible, visibleStateRef } = useIsVisible(root, visibleRef, true, 1000, 0)

    const hashRef = useRef("")

    useEffect(() => {
        if (!isVisible) {
            return
        }

        setImgData("")
        setLoaded("")
        setImgData(null)
        setLoadErr(false)

        const metaUrl = new URL(`${API_ENDPOINT}/media/${mediaId}`)
        metaUrl.searchParams.append("meta", "true")
        const thumbUrl = new URL(`${API_ENDPOINT}/media/${mediaId}`)
        thumbUrl.searchParams.append("thumbnail", "true")
        const fullresUrl = new URL(`${API_ENDPOINT}/media/${mediaId}`)
        fullresUrl.searchParams.append("fullres", "true")

        hashRef.current = mediaId

        const abortController = new AbortController();
        const signal = abortController.signal;

        setMetaPromise(_ => () => getImageMeta(metaUrl, mediaId, authHeader, signal, setLoadErr))
        setThumbPromise(_ => () => getImageData(thumbUrl, mediaId, authHeader, signal, setLoadErr))

        if (quality === "fullres" && !imgMeta?.mediaType.IsVideo) {
            setFullresPromise(_ => () => getImageData(fullresUrl, mediaId, authHeader, signal, setLoadErr))
        }

        // If the mediaId changes, we want to abort fetch requests
        // so we don't load a ton of images we don't need to
        return () => {
            abortController.abort()
        }

    }, [mediaId, isVisible])

    useEffect(() => {
        if (!metaPromise) {
            return
        }

        metaPromise().then((res) => {
            if (res && res.data && res.hash === hashRef.current && !loadError) {
                setImgMeta(res.data)
            }
        })
    }, [metaPromise, loadError])

    useEffect(() => {
        if (!thumbPromise) {
            return
        }

        thumbPromise().then((res) => {
            if (res && res.data && res.hash === hashRef.current) {
                setImgData(prev => {if (prev) {return prev}; return res.data})
                setLoaded(prev => {if (prev) {return prev}; return "thumbnail"})
            }
        })
    }, [thumbPromise, loadError])

    useEffect(() => {
        if (!fullresPromise) {
            return
        }

        fullresPromise().then((res) => {
            if (res && res.data && res.hash === hashRef.current) {
                setImgData(res.data)
                setLoaded("fullres")
            }
        })
    }, [fullresPromise, loadError])

    const sizer = useMemo(() => {
        if (!imgMeta) {
            return null
        }
        const sizer = imgMeta.mediaHeight > imgMeta.mediaWidth ? { width: '100%' } : { height: '100%' }
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
            {(quality === "fullres" && isVisible && loaded !== "fullres" && !loadError) && (
                <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
            )}

            <Image
                draggable={false}
                src={imgData}
                style={{ display: imgData && !loadError ? "" : "none", userSelect: 'none', minWidth: '100%', minHeight: '100%', ...sizer, ...imgStyle }}
            />

            {quality === "fullres" && imgMeta?.mediaType.IsVideo && (
                <video src="" controls/>
            )}

            {isVisible && blurhash && lazy && !imgData && (
                <Blurhash
                    style={{ position: "absolute" }}
                    height={containerStyle?.height ? containerStyle?.height : 300}
                    width={containerStyle?.width ? containerStyle?.width : 550}
                    hash={blurhash}
                />
            )}

        </ThumbnailContainer >
    )
}, (last, next) =>{
    if (last.mediaId !== next.mediaId) {
        return false
    }
    return true
})