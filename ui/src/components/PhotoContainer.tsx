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


export function useIsVisible(ref, maintained: boolean) {
    const [isIntersecting, setIntersecting] = useState(false);

    useEffect(() => {
        if (!ref.current) {
            return
        }
        let options = {
            rootMargin: "1000px"
        }
        const observer = new IntersectionObserver(([entry]) => {
            if (maintained && entry.isIntersecting) {
                setIntersecting(true)
            } else if (!maintained) {
                setIntersecting(entry.isIntersecting)
            }

        }, options)

        observer.observe(ref.current);
        return () => {
            observer.disconnect();
        };
    }, [ref]);

    return isIntersecting;
}

function getImageData(url, authHeader, setImgData, setImgLoaded, setLoadErr) {
    fetch(url, { headers: authHeader }).then(res => res.blob()).then((blob) => {
        if (blob.length === 0) {
            Promise.reject("Empty blob")
        }
        setImgData(URL.createObjectURL(blob))
        setImgLoaded(true)
    }).catch((r) => {
        setLoadErr(true)
        console.error("Failed to get image from server: ", r)
    })
}

export const MediaImage = ({
    mediaData,
    quality,
    lazy,
    containerStyle,
    imgStyle,
}: { mediaData: MediaData, quality: "thumbnail" | "fullres", lazy: boolean, containerStyle?: any, imgStyle?: MantineStyleProp }
) => {
    const [thumbLoaded, setThumbLoaded] = useState(false)
    const [fullresLoaded, setFullresLoaded] = useState(false)
    const [loadError, setLoadErr] = useState(false)
    const { authHeader } = useContext(userContext)
    const [thumbData, setThumbData] = useState("")
    const [fullresData, setFullresData] = useState("")

    const ref = useRef()
    const isVisible = useIsVisible(ref, true)

    const thumbUrl = new URL(`${API_ENDPOINT}/item/${mediaData.FileHash}`)
    thumbUrl.searchParams.append("thumbnail", "true")
    const fullresUrl = new URL(`${API_ENDPOINT}/item/${mediaData.FileHash}`)
    fullresUrl.searchParams.append("fullres", "true")

    useEffect(() => {
        setThumbData("")
        setFullresData("")
        setThumbLoaded(false)
        setFullresLoaded(false)
    }, [mediaData.FileHash])

    useEffect(() => {
        console.log(thumbUrl)
        if (!mediaData.FileHash) {
            console.error("LOAD ERR")
            setLoadErr(true)
        } else if (isVisible && !thumbData && !fullresData) {
            console.log("GETTIN")
            getImageData(thumbUrl, authHeader, setThumbData, setThumbLoaded, setLoadErr)
            if (quality === "fullres") {
                getImageData(fullresUrl, authHeader, setFullresData, setFullresLoaded, setLoadErr)
            }
        }
    }, [isVisible, thumbData, thumbUrl.toString()])

    if (quality === "fullres") {
        console.log(!lazy, isVisible, !fullresLoaded, !loadError)
    }

    return (
        <FlexColumnBox style={{ height: "100%", width: "100%" }}>
            <ThumbnailContainer reff={ref} style={containerStyle} >
                {(isVisible && loadError) && (
                    <IconExclamationCircle color="red" style={{ position: 'absolute' }} />
                )}

                {(!lazy && isVisible && !fullresLoaded && !loadError) && (
                    <Loader color="white" bottom={40} right={40} size={20} style={{ position: 'absolute' }} />
                )}

                {/* <AspectRatio ratio={mediaData.MediaWidth / mediaData.MediaHeight} maw={"100%"} mah={"100%"}> */}

                <Image
                    draggable={false}
                    src={fullresData ? fullresData : thumbData}
                    style={{ ...imgStyle, position: 'relative', display: (fullresData || thumbData) ? "block" : "none", userSelect: 'none', maxWidth: "100%" }}
                />

                {isVisible && mediaData.BlurHash && lazy && !thumbData && !fullresLoaded && (
                    <Blurhash
                        style={{ position: "absolute" }}
                        height={250}
                        width={550}
                        hash={mediaData.BlurHash}
                    />
                )}

                {/* </AspectRatio> */}

        </ThumbnailContainer >
        </FlexColumnBox>
    )
}