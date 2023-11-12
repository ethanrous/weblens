import { MediaImage } from '../components/PhotoContainer'
import { styled } from '@mui/joy'

export const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})

export const StyledLazyThumb = styled(MediaImage)({
    position: "absolute",

    objectFit: "cover",
    width: '100%',
    height: '100%',
    cursor: "pointer",
    overflow: "hidden",

    transitionDuration: "200ms",
    transform: "scale3d(1.00, 1.00, 1)",
    "&:hover": {
        transitionDuration: "200ms",
        transform: "scale3d(1.03, 1.03, 1)",
    }
})

