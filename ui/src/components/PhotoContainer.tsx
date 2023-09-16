import { useState, useEffect, useRef, useMemo } from "react";
import { Blurhash } from "react-blurhash";
import ReactPlayer from 'react-player'

import Box from '@mui/material/Box';
import { CircularProgress } from '@mui/material'
import styled from "@emotion/styled";

import { MediaData } from '../types/Generic'
import { fetchThumb64 } from '../api/ApiFetch'

export const MediaFullresComponent = ({ mediaData }: { mediaData: MediaData }) => {
    const [fullResLoaded, setFullResLoaded] = useState(false)
    const [fullRes64, setFullRes64] = useState("")
    const [fullresUrlObj, setFullresUrlObj] = useState("")

    const vidRef = useRef(null)
    const hashRef = useRef(mediaData.FileHash)

    useEffect(() => {
        const baseUrl = "http://localhost:3000/api" //item/${mediaData.FileHash}`
        setFullResLoaded(false)
        setFullRes64("")
        hashRef.current = mediaData.FileHash

        if (mediaData.MediaType.IsVideo) {
            let fullresUrlObj = new URL(`${baseUrl}/stream/${mediaData.FileHash}`)
            setFullresUrlObj(fullresUrlObj.toString())

        } else {
            let fullresUrlObj = new URL(`${baseUrl}/item/${mediaData.FileHash}`)
            fullresUrlObj.searchParams.append('fullres', 'true')

            fetch(fullresUrlObj.toString()).then(res => res.blob()).then(blob => new Promise((resolve, reject) => {
                const reader = new FileReader()
                reader.onloadend = () => resolve(reader.result)
                reader.onerror = reject
                reader.readAsDataURL(blob)
            })).then((str: string) => { if (hashRef.current != mediaData.FileHash) { return }; setFullRes64(str) })

        }

    }, [mediaData.FileHash])

    return (
        <Box display={"flex"} alignItems={"center"} position={"relative"} justifyContent={"center"} height={"100%"} width={"100%"}>
            {!fullResLoaded && mediaData && ( //!mediaData.MediaType.IsVideo && (
                <MediaThumbComponent fileHash={mediaData.FileHash} blurhash={""} />
                // <img
                //     src={thumb64}
                //     height={"100%"}
                //     width={"100%"}
                //     style={{ objectFit: "contain", position: "static", display: fullResLoaded ? "none" : "block", objectPosition: "center" }}
                // />
            )}
            {!fullResLoaded && (
                <CircularProgress size={20} color={"inherit"} style={{ bottom: 10, right: 10, position: "absolute", alignSelf: "right", zIndex: 101 }} />
            )}
            {!mediaData.MediaType.IsVideo && (
                <img
                    src={fullRes64}
                    height={"100%"}
                    width={"100%"}
                    onLoad={() => { setFullResLoaded(true) }}
                    style={{ objectFit: "contain", position: "absolute", display: fullResLoaded ? "block" : "none" }}
                />
            )}
            {mediaData.MediaType.IsVideo && (
                <Box position={"absolute"} height={"100%"} width={"100%"}>
                    <ReactPlayer
                        height={"100%"}
                        width={"100%"}
                        onReady={() => console.log("IM READY")}
                        style={{ position: "absolute" }}
                        volume={0}
                        ref={vidRef}
                        onPlay={() => setFullResLoaded(true)}
                        url={fullresUrlObj}
                        playing={true}
                        onSeek={(e) => console.log(e)}
                        onBuffer={() => console.log("Bufferin")}
                    />
                </Box>
            )}
        </Box>
    )
}

const ThumbnailContainer = styled(Box)({
    height: "100%",
    width: "100%",
    display: "flex",
    justifyContent: "center",
    overflow: "hidden",
    objectFit: "contain"
})

export function useIsVisible(ref) {
    const [isIntersecting, setIntersecting] = useState(false);

    useEffect(() => {
        let options = {
            rootMargin: "1000px"
        }
        const observer = new IntersectionObserver(([entry]) => {
            setIntersecting(entry.isIntersecting)
        }, options)

        observer.observe(ref.current);
        return () => {
            observer.disconnect();
        };
    }, [ref]);

    return isIntersecting;
}

const StyledImg = styled('img')({
    height: "100%",
    width: "100%",
    loading: "lazy",
    objectFit: "contain"
})

export const MediaThumbComponent = ({
    fileHash,
    blurhash,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)
    const [imageData, setImageData] = useState("")
    const ref = useRef()
    const isVisible = useIsVisible(ref)
    useEffect(() => {
        if (isVisible && imageData == "") {
            fetchThumb64(fileHash, setImageData)
        }
    }, [isVisible])

    return (
        <ThumbnailContainer ref={ref} >
            <StyledImg
                {...props}
                src={imageData}
                style={{ position: "absolute", display: imageData ? "block" : "none" }}

                //loading="lazy" // WHY DONT THIS WORK
                onLoad={() => { setImageLoaded(true) }}
            />
            {blurhash && !imageLoaded && (
                <Blurhash
                    style={{ position: "absolute", display: isVisible ? "block" : "none" }}
                    height={250}
                    width={475}
                    hash={blurhash}
                />

            )}
        </ThumbnailContainer>
    )
}
