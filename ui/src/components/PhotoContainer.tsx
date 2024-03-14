import { useState, useEffect, useRef, useContext, memo, useCallback } from "react";
import { Blurhash } from "react-blurhash";
import { userContext } from "../Context";
import { IconExclamationCircle, IconPhoto } from "@tabler/icons-react"
import { Box, CSSProperties, Loader } from "@mantine/core";
import API_ENDPOINT from '../api/ApiEndpoint'
import { MediaData } from "../types/Types";
import './style.css'

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
    setMediaCallback,
    quality,
    pageNumber = undefined,
    lazy = true,
    expectFailure = false,
    preventClick = false,
    doFetch = true,
    containerStyle,
    disabled = false,
}: { media: MediaData, setMediaCallback?: (mediaId: string, quality: "thumbnail" | "fullres", data: ArrayBuffer) => void, quality: "thumbnail" | "fullres", pageNumber?: number, lazy?: boolean, expectFailure?: boolean, preventClick?: boolean, doFetch?: boolean, containerStyle?: CSSProperties, disabled?: boolean }
) => {
    const [loaded, setLoaded] = useState(media?.fullres ? "fullres" : media?.thumbnail ? "thumbnail" : "")
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [imgData, setImgData] = useState(null)
    const visibleRef = useRef(null)
    const hashRef = useRef("")
    const abortController = new AbortController();

    const fetchFullres = useCallback(async () => {
        const url = new URL(`${API_ENDPOINT}/media/${media.fileHash}/fullres`)
        if (pageNumber !== undefined) {
            url.searchParams.append("page", pageNumber.toString())
        }
        const ret = await getImageData(url.toString(), media.fileHash, authHeader, abortController.signal, setLoadErr)
        if (!ret) {
            return
        }
        if (setMediaCallback) {
            setMediaCallback(media.fileHash, 'fullres', ret.data)
        } else {
            media.fullres = ret.data
        }
        setImgData(URL.createObjectURL(new Blob([ret.data])))
        setLoaded("fullres")
    }, [media?.fileHash, authHeader])

    const fetchThumbnail = useCallback(async () => {
        const ret = await getImageData(`${API_ENDPOINT}/media/${media.fileHash}/thumbnail`, media.fileHash, authHeader, abortController.signal, setLoadErr)
        if (!ret) {
            return
        }
        if (setMediaCallback) {
            setMediaCallback(media.fileHash, 'thumbnail', ret.data)
        } else {
            media.thumbnail = ret.data
        }
        setImgData(prev => prev === "" ? URL.createObjectURL(new Blob([ret.data])) : prev)
        setLoaded(prev => prev === "" ? "thumbnail" : prev)
    }, [media?.fileHash, authHeader])

    useEffect(() => {
        setLoadErr(false)
        if (!media?.fileHash) {
            return
        }
        hashRef.current = media.fileHash
        setLoaded("")

        if (media.fullres && quality === "fullres") {
            setImgData(URL.createObjectURL(new Blob([media.fullres])))
            setLoaded("fullres")
        } else if (media.thumbnail) {
            setImgData(URL.createObjectURL(new Blob([media.thumbnail])))
            setLoaded("thumbnail")
        } else {
            setImgData("")
        }

        if (!media.fullres && doFetch && quality === "fullres") {
            fetchFullres()
        }

        if (!media.thumbnail && doFetch) {
            fetchThumbnail()
        }

        return () => abortController.abort()
    }, [media?.fileHash])

    return (
        <Box
            className="photo-container"
            ref={visibleRef}
            style={{ ...containerStyle }}
            onDrag={(e) => { e.preventDefault(); e.stopPropagation() }}
            onClick={e => { preventClick && e.stopPropagation() }}
        >
            {(loadError && !expectFailure) && (
                <IconExclamationCircle color="red" />
            )}
            {((loadError && expectFailure) || !media?.fileHash) && (
                <IconPhoto />
            )}
            {(quality === "fullres" && loaded !== "fullres" && !loadError) && (
                <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
            )}

            <img
                alt=""
                className={quality === 'thumbnail' ? "media-thumbnail" : "media-fullres"}
                draggable={false}
                src={imgData}
                style={{ display: imgData && !loadError ? "" : "none", filter: disabled ? "grayscale(100%)" : '' }}
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
        </Box>
    )
}, (last, next) => {
    if (last.media?.fileHash !== next.media?.fileHash) {
        return false
    } else if (last.containerStyle?.height !== next.containerStyle?.height) {
        return false
    } else if (last.media?.thumbnail !== next.media?.thumbnail) {
        return false
    } else if (last.media?.fullres !== next.media?.fullres) {
        return false
    }

    return true
})