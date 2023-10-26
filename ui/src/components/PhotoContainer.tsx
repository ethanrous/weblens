import { useState, useEffect, useRef, useMemo, ComponentProps } from "react";
import { Blurhash } from "react-blurhash";
import API_ENDPOINT from '../api/ApiEndpoint'

import Box from '@mui/material/Box';
import { CircularProgress } from '@mui/material'
import styled from "@emotion/styled";
import { useCookies } from "react-cookie";
import { useNavigate } from "react-router-dom";

// Styles

const ThumbnailContainer = styled(Box)({
    top: 0,
    left: 0,
    height: "100%",
    width: "100%",
    display: "flex",
    position: "absolute",
    justifyContent: "center",
    overflow: "hidden",
    objectFit: "contain"
})

const StyledLoader = styled(CircularProgress)({
    position: "absolute",
    zIndex: 1,
    bottom: "10px",
    right: "10px",
    color: "rgb(255, 255, 255)"
})

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
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token'])
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
        if (isVisible && !imgData) {
            fetch(imgUrl, { headers: { "Authorization": `${cookies['weblens-username']},${cookies['weblens-login-token']}` } }).then(res => res.blob()).then((blob) => {
                setImgData(URL.createObjectURL(blob))
                setImageLoaded(true)
            })
        }

    }, [isVisible, imgUrl])

    return (
        <ThumbnailContainer ref={ref} >
            {!imageLoaded && (
                <StyledLoader size={20} />
            )}

            <img
                height={"100%"}
                width={"100%"}
                src={imgData}
                crossOrigin="use-credentials"

                {...props}

                style={{ position: "absolute", display: imageLoaded ? "block" : "none" }}
            />
            {mediaData.BlurHash && lazy && !imageLoaded && (
                <Blurhash
                    style={{ position: "absolute" }}
                    height={250}
                    width={550}
                    hash={mediaData.BlurHash}
                />

            )}
        </ThumbnailContainer>
    )
}