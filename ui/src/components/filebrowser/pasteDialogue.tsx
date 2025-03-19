import { PresentationContainer } from '@weblens/components/Presentation'
import WeblensButton from '@weblens/lib/WeblensButton'
import { uploadViaUrl } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensMedia, { PhotoQuality } from '@weblens/types/media/Media'
import { MediaImage } from '@weblens/types/media/PhotoContainer'

function PasteDialogue() {
    const filesMap = useFileBrowserStore((state) => state.filesMap)
    const contentId = useFileBrowserStore((state) => state.contentId)
    const shareId = useFileBrowserStore((state) => state.shareId)
    const pasteImage = useFileBrowserStore((state) => state.pasteImgBytes)
    const pasteImgBytes = useFileBrowserStore((state) => state.pasteImgBytes)

    const setPasteImgBytes = useFileBrowserStore(
        (state) => state.setPasteImgBytes
    )

    if (!pasteImgBytes) {
        return null
    }

    const media = new WeblensMedia({ contentId: 'paste' })
    media.SetThumbnailBytes(pasteImage)

    return (
        <PresentationContainer
            onClick={() => {
                setPasteImgBytes(null)
            }}
        >
            <div className="absolute z-2 flex h-full w-full flex-col items-center justify-center">
                <p className="pb-[50px] text-[40px] font-bold">
                    Upload from clipboard?
                </p>
                <div
                    className="bg-bottom-grey h-1/2 w-max overflow-hidden rounded-lg p-3"
                    onClick={(e) => {
                        e.stopPropagation()
                    }}
                >
                    <MediaImage media={media} quality={PhotoQuality.LowRes} />
                </div>
                <div className="flex w-[50%] flex-row justify-between gap-6">
                    <WeblensButton
                        label={'Cancel'}
                        squareSize={50}
                        fillWidth
                        subtle
                        onClick={(e) => {
                            e.stopPropagation()

                            setPasteImgBytes(null)
                        }}
                    />
                    <WeblensButton
                        label={'Upload'}
                        squareSize={50}
                        fillWidth
                        onClick={(e) => {
                            e.stopPropagation()
                            uploadViaUrl(
                                pasteImage,
                                contentId,
                                filesMap,
                                shareId
                            )
                                .then(() => setPasteImgBytes(null))
                                .catch(ErrorHandler)
                        }}
                    />
                </div>
            </div>
        </PresentationContainer>
    )
}
export default PasteDialogue
