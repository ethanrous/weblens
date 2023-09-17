import { useState, useEffect, useRef, useMemo, ComponentProps } from "react";
import { Blurhash } from "react-blurhash";
import ReactPlayer from 'react-player'

import Box from '@mui/material/Box';
import { CircularProgress } from '@mui/material'
import styled from "@emotion/styled";

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


export const MediaImage = ({
    mediaData,
    quality,
    lazy,
    ...props
}) => {
    const [imageLoaded, setImageLoaded] = useState(false)
    const ref = useRef()
    const isVisible = useIsVisible(ref)

    const imgUrl = new URL(`http://localhost:3000/api/item/${mediaData.FileHash}`)
    imgUrl.searchParams.append(quality, "true")

    return (
        <ThumbnailContainer ref={ref} >
            {!imageLoaded && (
                <StyledLoader size={20} />
            )}
            <img
                height={"100%"}
                width={"100%"}
                loading={lazy ? "lazy" : "eager"}

                {...props}

                src={imgUrl.toString()}
                style={{ position: "absolute", display: isVisible ? "block" : "none" }}

                onLoad={() => { setImageLoaded(true) }}
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

// Component

// export const MediaFullresComponent = ({ mediaData, dispatch }: { mediaData: MediaData, dispatch: React.Dispatch<any> }) => {
//     const [fullResLoaded, setFullResLoaded] = useState(false)
//     const [fullresUrlObj, setFullresUrlObj] = useState("")

//     if (!mediaData.MediaType.IsVideo && mediaData.Fullres64 == null) {
//         fetchFullres64(mediaData.FileHash, (fullres64: string) => dispatch({ type: "insert_fullres", hash: mediaData.FileHash, fullres64: fullres64 }))
//     } else if (fullResLoaded != true) {
//         setFullResLoaded(true)
//     }

//     const vidRef = useRef(null)
//     const hashRef = useRef(mediaData.FileHash)

//     useEffect(() => {
//         const baseUrl = "http://localhost:3000/api" //item/${mediaData.FileHash}`
//         console.log("Setting false")
//         setFullResLoaded(false)
//         hashRef.current = mediaData.FileHash

//         if (mediaData.MediaType.IsVideo) {
//             let fullresUrlObj = new URL(`${baseUrl}/stream/${mediaData.FileHash}`)
//             // setFullresUrlObj(fullresUrlObj.toString())
//         }

//     }, [mediaData.FileHash])

//     return (
//         <Box display={"flex"} alignItems={"center"} position={"relative"} justifyContent={"center"} height={"100%"} width={"100%"}>
//             {!fullResLoaded && mediaData && (
//                 <FullscreenThumb mediaData={mediaData} showBlurhash={false} dispatch={dispatch} />
//             )}
//             {!fullResLoaded && (
//                 <CircularProgress size={20} color={"inherit"} style={{ bottom: 10, right: 10, position: "absolute", alignSelf: "right", zIndex: 101 }} />
//             )}
//             {fullResLoaded && !mediaData.MediaType.IsVideo && (
//                 <img
//                     src={mediaData.Fullres64}
//                     height={"100%"}
//                     width={"100%"}
//                     //onLoad={() => { setFullResLoaded(true) }}
//                     style={{ objectFit: "contain", position: "absolute", display: fullResLoaded ? "block" : "none" }}
//                 />
//             )}
//             {mediaData.MediaType.IsVideo && (
//                 <Box position={"absolute"} height={"100%"} width={"100%"}>
//                     <ReactPlayer
//                         height={"100%"}
//                         width={"100%"}
//                         style={{ position: "absolute" }}
//                         volume={0}
//                         ref={vidRef}
//                         onPlay={() => setFullResLoaded(true)}
//                         url={fullresUrlObj}
//                         playing={true}
//                         onSeek={(e) => console.log(e)}
//                         onBuffer={() => console.log("Bufferin")}
//                     />
//                 </Box>
//             )}
//         </Box>
//     )
// }
