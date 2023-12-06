import { useState, useEffect, useRef, useContext } from "react";
import { Blurhash } from "react-blurhash";
import API_ENDPOINT from '../api/ApiEndpoint'

import { userContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react"
import { MediaData } from "../types/Types";
import { AspectRatio, Box, Image, Loader, MantineStyleProp } from "@mantine/core";
import { FlexColumnBox } from "../Pages/FileBrowser/FilebrowserStyles";

// Styles

const ThumbnailContainer = ({ reff, style, children }) => {
    return (
        <Box
            ref={reff}
            draggable={false}
            style={{
                ...style,
            // height: 'max-content',
                height: '100%',
                // width: 'max-content',
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                // position: 'absolute'
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

export const MediaImage = ({
    mediaData,
    quality,
    lazy,
    containerStyle,
    imgStyle,
    root
}: { mediaData: MediaData, quality: "thumbnail" | "fullres", lazy: boolean, containerStyle?: any, imgStyle?: MantineStyleProp, root?}
) => {
    const [loaded, setLoaded] = useState(false)
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [imgData, setImgData] = useState(null)

    const [thumbPromise, setThumbPromise] = useState(null)
    const [fullresPromise, setFullresPromise] = useState(null)

    const visibleRef = useRef()
    const { isVisible, visibleStateRef } = useIsVisible(root, visibleRef, false, 1000, 0)

    const hashRef = useRef("")

    const abortController = new AbortController();
    const signal = abortController.signal;

    const thumbUrl = new URL(`${API_ENDPOINT}/item/${mediaData.FileHash}`)
    thumbUrl.searchParams.append("thumbnail", "true")
    const fullresUrl = new URL(`${API_ENDPOINT}/item/${mediaData.FileHash}`)
    fullresUrl.searchParams.append("fullres", "true")

    useEffect(() => {
        setImgData(null)
        setThumbPromise(null)
        setFullresPromise(null)
        setLoaded(false)
        hashRef.current = mediaData.FileHash
    }, [mediaData.FileHash])

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
            setThumbPromise(getImageData(thumbUrl, mediaData.FileHash, authHeader, signal))
            if (quality === "fullres") {
                setFullresPromise(getImageData(fullresUrl, mediaData.FileHash, authHeader, signal))
            }
        }
        return () => {
            if (!visibleStateRef.current || mediaData.FileHash !== hashRef.current) {
                abortController.abort()
            }
        }
    }, [isVisible, hashRef.current])

    return (
        <FlexColumnBox style={{ height: "100%", width: "100%" }}>
            <ThumbnailContainer reff={visibleRef} style={containerStyle} >
                {(isVisible && loadError) && (
                    <IconExclamationCircle color="red" style={{ position: 'absolute' }} />
                )}

                {(!lazy && isVisible && !loaded && !loadError) && (
                    <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
                )}

                <Image
                    draggable={false}
                    src={imgData}
                    style={{ position: 'relative', display: imgData ? "block" : "none", userSelect: 'none', ...imgStyle }}
                />

                {isVisible && mediaData.BlurHash && lazy && !imgData && (
                    <Blurhash
                        style={{ position: "absolute" }}
                        height={250}
                        width={550}
                        hash={mediaData.BlurHash}
                    />
                )}

        </ThumbnailContainer >
        </FlexColumnBox>
    )
}