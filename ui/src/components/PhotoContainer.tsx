import { useState, useEffect, useRef, useContext } from "react";
import { Blurhash } from "react-blurhash";
import API_ENDPOINT from '../api/ApiEndpoint'

import Box from '@mui/material/Box';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { CircularProgress } from '@mui/material'
import { userContext } from "../Context";

// Styles

const ThumbnailContainer = ({ reff, sx, ...props }) => {
    return (
        <Box
            ref={reff}
            sx={{
                ...sx,
                height: '100%',
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                position: 'absolute'
            }}
            {...props}
        />
    )
}


const StyledLoader = ({ loading, error }) => {
    if (!loading && !error) {
        return null
    } else if (loading && !error) {
        return (
            <CircularProgress size={20} sx={{
                position: "absolute",
                zIndex: 1,
                top: "10px",
                right: "10px",
                color: "rgb(255, 255, 255)"
            }} />
        )
    } else {
        return (
            <ErrorOutlineIcon sx={{
                position: "absolute",
                zIndex: 1,
                top: "10px",
                right: "10px",
                color: "rgb(255, 51, 51)"
            }} />
        )
    }
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


export const MediaImage = ({
    mediaData,
    quality,
    lazy,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)
    const [loadError, setLoadError] = useState(false)
    const { authHeader, userInfo } = useContext(userContext)
    const ref = useRef()
    const isVisible = useIsVisible(ref, true)
    const [imgData, setImgData] = useState("")
    const [imgUrl, setImgUrl] = useState("")

    useEffect(() => {
        const url = new URL(`${API_ENDPOINT}/item/${mediaData.FileHash}`)
        url.searchParams.append(quality, "true")
        setImgUrl(url.toString())
    }, [])

    useEffect(() => {
        if (!mediaData.FileHash) {
            setLoadError(true)
        } else if (isVisible && !imgData) {
            fetch(imgUrl, { headers: authHeader }).then(res => res.blob()).then((blob) => {
                if (blob.length === 0) {
                    Promise.reject("Empty blob")
                }
                setImgData(URL.createObjectURL(blob))
                setImageLoaded(true)
            }).catch((r) => {
                console.error("Failed to get image from server: ", r)
            })
        }
    }, [isVisible, imgUrl])

    return (
        <ThumbnailContainer reff={ref} sx={props.sx} onDrag={(e) => { console.log(e); e.preventDefault(); e.stopPropagation() }}>
            <StyledLoader loading={!imageLoaded} error={loadError} />
            <img
                draggable={false}
                height={"max-content"}
                width={"100%"}
                src={imgData}
                crossOrigin="use-credentials"
                // onDrag={(e) => { console.log(e); e.preventDefault(); e.stopPropagation() }}

                {...props}

                style={{ display: imageLoaded ? "block" : "none" }}
            />
            {mediaData.BlurHash && lazy && !imageLoaded && (
                <Blurhash
                    style={{ position: "absolute" }}
                    height={250}
                    width={550}
                    hash={mediaData.BlurHash}
                />

            )}
        </ThumbnailContainer >
    )
}