import { useState, useEffect, useRef, useContext, memo, useCallback } from "react";
import { Blurhash } from "react-blurhash";
import { userContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react"
import { Box, CSSProperties, Image, Loader, MantineStyleProp } from "@mantine/core";


import API_ENDPOINT from '../api/ApiEndpoint'
import { MediaData } from "../types/Types";
import './style.css'

// Styles

const ThumbnailContainer = ({ reff, style, children }) => {
    return (
        <Box
            ref={reff}
            draggable={false}
            style={{
                display: 'flex',
                height: '100%',
                width: '100%',
                alignItems: 'center',
                justifyContent: 'center',
                ...style,
            }}
            onDrag={(e) => { e.preventDefault(); e.stopPropagation() }}
            children={children}
        />
    )
}

function getImageData(url, filehash, authHeader, signal, setLoadErr) {
    const res = fetch(url, { headers: authHeader, signal })
        .then(res => {
            if (res.status !== 200) {
                return Promise.reject(res.statusText)
            }
            return res.arrayBuffer()
        })
        .then(buf => {
            if (buf.byteLength === 0) {
                Promise.reject("Empty blob")
            }

            return { data: buf, hash: filehash }
        })
        .catch((r) => {
            if (!signal.aborted) {
                console.error("Failed to get image from server:", r)
                setLoadErr(true)
            }
        })

    return res
}

export const MediaImage = memo(({
    media,
    quality,
    lazy = true,
    expectFailure = false,
    containerStyle,
    imgStyle,
}: { media: MediaData, quality: "thumbnail" | "fullres", lazy?: boolean, expectFailure?: boolean, containerStyle?: CSSProperties, imgStyle?: MantineStyleProp }
) => {
    const [loaded, setLoaded] = useState(media?.fullres ? "fullres" : media?.thumbnail ? "thumbnail" : "")
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [imgData, setImgData] = useState(null)
    const [innerImgStyle, setImgStyle] = useState(imgStyle)
    const visibleRef = useRef(null)
    const hashRef = useRef("")
    const abortController = new AbortController();

    const fetchFullres = useCallback(async () => {
        const ret = await getImageData(`${API_ENDPOINT}/media/${media.fileHash}?fullres=true`, media.fileHash, authHeader, abortController.signal, setLoadErr)
        if (!ret) {
            return
        }
        media.fullres = ret.data
        setImgData(URL.createObjectURL(new Blob([ret.data])))
        setLoaded("fullres")
        setImgStyle(imgStyle)
    }, [media?.fileHash, authHeader])

    const fetchThumbnail = useCallback(async () => {
        const ret = await getImageData(`${API_ENDPOINT}/media/${media.fileHash}?thumbnail=true`, media.fileHash, authHeader, abortController.signal, setLoadErr)
        if (!ret) {
            return
        }
        media.thumbnail = ret.data
        setImgData(prev => prev === "" ? URL.createObjectURL(new Blob([ret.data])) : prev)
        setLoaded(prev => prev === "" ? "thumbnail" : prev)
        setImgStyle(imgStyle)
    }, [media?.fileHash, authHeader])

    useEffect(() => {
        if (!media || !media.fileHash) {
            return
        }
        hashRef.current = media.fileHash
        setLoaded("")

        if (media.fullres && quality === "fullres") {
            setImgData(URL.createObjectURL(new Blob([media.fullres])))
            setLoaded("fullres")
            setImgStyle(imgStyle)
        } else if (media.thumbnail) {
            setImgData(URL.createObjectURL(new Blob([media.thumbnail])))
            setLoaded("thumbnail")
            setImgStyle(imgStyle)
        } else {
            setImgData("")
        }

        if (!media.fullres && quality === "fullres") {
            fetchFullres()
        }
        if (!media.thumbnail) {
            fetchThumbnail()
        }

        return () => abortController.abort()
    }, [media?.fileHash])

    const imgClass = quality === 'thumbnail' ? "media-image" : ""

    return (
        <ThumbnailContainer reff={visibleRef} style={containerStyle}>
            {(loadError && !expectFailure) && (
                <IconExclamationCircle color="red" style={{ position: 'absolute' }} />
            )}
            {(loadError && expectFailure) && (
                <IconPhoto style={{ position: 'absolute' }} />
            )}
            {(quality === "fullres" && loaded !== "fullres" && !loadError) && (
                <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
            )}

            <Image
                className={imgClass}
                draggable={false}
                src={imgData}
                style={{ display: imgData && !loadError ? "" : "none", userSelect: 'none', flex: 'none', ...innerImgStyle }}
            />

            {quality === "fullres" && media?.mediaType?.IsVideo && (
                <video src="" controls />
            )}

            {media?.blurHash && lazy && !imgData && (
                <Blurhash
                    style={{ position: "absolute" }}
                    height={visibleRef?.current?.clientHeight ? visibleRef.current.clientHeight : 0}
                    width={visibleRef?.current?.clientWidth ? visibleRef.current.clientWidth : 0}
                    hash={media.blurHash}
                />
            )}
            {!media?.blurHash && lazy && !imgData && !loadError && (
                <IconPhoto />
            )}

        </ThumbnailContainer >
    )
}, (last, next) => {
    if (last.media?.fileHash !== next.media?.fileHash) {
        return false
    } else if (last.containerStyle?.height !== next.containerStyle?.height) {
        return false
    } else if (last.imgStyle !== next.imgStyle) {
        return false
    } else if (last.media?.thumbnail !== next.media?.thumbnail) {
        return false
    } else if (last.media?.fullres !== next.media?.fullres) {
        return false
    }

    return true
})