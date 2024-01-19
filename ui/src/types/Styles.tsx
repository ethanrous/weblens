import { useRef } from 'react'
import { MediaImage } from '../components/PhotoContainer'
import { MediaData } from './Types'
import { Box, MantineStyleProp, ScrollArea } from '@mantine/core'

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

export const FilesWrapper = ({ size = 300, gap = 20, reff, children, style }: { size?: number, gap?: number, reff?, children, style?: MantineStyleProp }) => {
    const boxRef = useRef(null)
    // width

    return (
        <ScrollArea ref={reff} type='never' style={{ width: '100%', height: '100%', borderRadius: '10px' }}>
            <Box
                ref={boxRef}
                children={children}
                style={{
                    display: 'grid',
                    // flexWrap: 'wrap',
                    gridGap: 16,
                    gridTemplateColumns: "repeat(auto-fill,minmax(220px,1fr))",
                    paddingLeft: 2,
                    paddingTop: 2,
                    paddingRight: '20px',
                    width: "100%",
                    ...style
                }}
            />
        </ScrollArea>
    )
}
