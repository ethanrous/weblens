import { MediaImage } from '../components/PhotoContainer'
import { styled } from '@mui/joy'

export const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})

export const StyledLazyThumb = ({ mediaData, quality, lazy }) => {
    return (
        <MediaImage mediaData={mediaData} quality={quality} lazy={lazy} containerStyle={{
            // position: "absolute",

            width: '100%',

            cursor: "pointer",
            overflow: "hidden",

            transitionDuration: "200ms",
            transform: "scale3d(1.00, 1.00, 1)",
            "&:hover": {
                transitionDuration: "200ms",
                transform: "scale3d(1.03, 1.03, 1)",
            }
        }}
            imgStyle={{
                objectFit: "cover",
                height: '100%',
            }}
        />
    )
}

