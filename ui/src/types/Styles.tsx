import { MediaImage } from '../components/PhotoContainer'
import { styled } from '@mui/joy'
import { MediaData } from './Types'

export const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})

export const StyledLazyThumb = ({ mediaData, quality, lazy, root }: { mediaData: MediaData, quality, lazy, root }) => {
    let sizer
    if (mediaData.MediaHeight < mediaData.MediaWidth) {
        sizer = { height: "100%" }
    } else {
        sizer = { width: "100%" }
    }
    return (
        <MediaImage mediaData={mediaData} quality={quality} lazy={lazy}
            imgStyle={{
                transitionDuration: "200ms",
                transform: "scale3d(1.00, 1.00, 1)",
                "&:hover": {
                    transitionDuration: "200ms",
                    transform: "scale3d(1.03, 1.03, 1)",
                },
                objectFit: "cover",
                ...sizer,
            }}
            root={root}
        />
    )
}

