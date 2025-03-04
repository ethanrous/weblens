import {
    IconFile,
    IconFileZip,
    IconFolder,
    IconPhoto,
} from '@tabler/icons-react'
import { useResize } from '@weblens/components/hooks'
import WeblensFile from '@weblens/types/files/File'
import { PhotoQuality } from '@weblens/types/media/Media'
import { useMediaStore } from '@weblens/types/media/MediaStateControl'
import { MediaImage } from '@weblens/types/media/PhotoContainer'
import { useState } from 'react'

function FileVisual({
    file,
    allowMedia = false,
}: {
    file: WeblensFile
    allowMedia?: boolean
}) {
    const [containerRef, setContainerRef] = useState<HTMLDivElement>(null)
    const containerSize = useResize(containerRef)
    const mediaData = useMediaStore((state) =>
        state.mediaMap.get(file.GetContentId())
    )

    if (!file) {
        return null
    }

    if (file.IsFolder()) {
        if (file.GetContentId() !== '') {
            const containerQuanta = Math.ceil(containerSize.height / 100)
            return (
                <div
                    ref={setContainerRef}
                    className="relative flex h-full w-full items-center justify-center"
                >
                    <div
                        className="relative z-20 h-[90%] w-[90%]"
                        style={{
                            translate: `${containerQuanta * -3}px ${containerQuanta * -3}px`,
                        }}
                    >
                        <MediaImage
                            media={mediaData}
                            quality={PhotoQuality.LowRes}
                        />
                    </div>
                    <div className="bg-wl-outline-subtle outline-theme-text absolute z-10 h-[88%] w-[88%] rounded opacity-75 outline outline-2" />
                    <div
                        className="bg-wl-outline-subtle outline-theme-text absolute h-[88%] w-[88%] rounded opacity-50 outline outline-2"
                        style={{
                            translate: `${containerQuanta * 3}px ${containerQuanta * 3}px`,
                        }}
                    />
                </div>
            )
        } else {
            return (
                <IconFolder
                    stroke={1}
                    className="z-10 h-3/4 w-3/4 shrink-0 text-[--wl-file-text-color]"
                />
            )
        }
    }

    if (mediaData && (!mediaData.IsImported() || !allowMedia)) {
        return <IconPhoto stroke={1} className="shrink-0" />
    } else if (mediaData && allowMedia && mediaData.IsImported()) {
        return <MediaImage media={mediaData} quality={PhotoQuality.LowRes} />
    }

    const extIndex = file.GetFilename().lastIndexOf('.')
    const ext = file
        .GetFilename()
        .slice(extIndex + 1, file.GetFilename().length)
    const textSize = `${Math.floor(containerSize?.width / (ext.length + 5))}px`

    switch (ext) {
        case 'zip':
            return <IconFileZip />
        default:
            return (
                <div
                    ref={setContainerRef}
                    className="flex h-full w-full items-center justify-center"
                >
                    <IconFile stroke={1} className="h-3/4 w-3/4" />
                    {extIndex !== -1 && (
                        <p
                            className="absolute select-none font-semibold"
                            style={{ fontSize: textSize }}
                        >
                            .{ext}
                        </p>
                    )}
                </div>
            )
    }
}

export default FileVisual
