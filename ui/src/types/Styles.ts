import { MediaThumbComponent } from '../components/PhotoContainer'
import styled from '@emotion/styled'

export const StyledLazyThumb = styled(MediaThumbComponent)({
    position: "relative",

    //objectFit: "cover",
    cursor: "pointer",
    overflow: "hidden",

    transitionDuration: "200ms",
    transform: "scale3d(1.00, 1.00, 1)",
    "&:hover": {
        transitionDuration: "200ms",
        transform: "scale3d(1.03, 1.03, 1)",
    }
})
