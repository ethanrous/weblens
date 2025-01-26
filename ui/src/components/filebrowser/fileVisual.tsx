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
                    className="relative flex w-full h-full justify-center items-center "
                >
                    <div
                        className="relative w-[90%] h-[90%] z-20"
                        style={{
                            translate: `${containerQuanta * -3}px ${containerQuanta * -3}px`,
                        }}
                    >
                        <MediaImage
                            media={mediaData}
                            quality={PhotoQuality.LowRes}
                        />
                    </div>
                    <div className="absolute w-[88%] h-[88%] bg-wl-outline-subtle outline outline-2 outline-theme-text opacity-75 rounded z-10" />
                    <div
                        className="absolute w-[88%] h-[88%] bg-wl-outline-subtle outline outline-2 outline-theme-text opacity-50 rounded"
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
                    className="h-3/4 w-3/4 z-10 shrink-0 text-[--wl-file-text-color]"
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
                    className="flex justify-center items-center w-full h-full"
                >
                    <IconFile stroke={1} className="w-3/4 h-3/4" />
                    {extIndex !== -1 && (
                        <p
                            className="font-semibold absolute select-none"
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
