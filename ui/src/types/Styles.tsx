import { MediaImage } from '../components/PhotoContainer'
import { MediaData } from './Types'
import { Box, ScrollArea } from '@mantine/core'

export const BlankCard = ({ scale }) => <Box style={{ height: scale, flexGrow: 9999 }} />

export const StyledLazyThumb = ({ mediaData, quality, lazy, root }: { mediaData: MediaData, quality, lazy, root }) => {
    return (
        <MediaImage mediaId={mediaData.fileHash} blurhash={mediaData.blurHash} quality={quality} lazy={lazy}
            imgStyle={{
                objectFit: "cover",
            }}
            root={root}
        />
    )
}

export const ItemsWrapper = ({ size = 200, reff, children }: { size?, reff?, children }) => {
    return (

        <ScrollArea type='never' style={{ width: '100%', borderRadius: '10px' }}>
            <Box
                ref={reff}
                children={children}
                style={{
                    display: 'grid',
                    gridGap: '16px',
                    gridTemplateColumns: `repeat(auto-fill, ${size}px)`,
                    padding: '10px',
                    paddingRight: "1vw",
                    width: "100%"
                }}
            />
        </ScrollArea>
    )
}
