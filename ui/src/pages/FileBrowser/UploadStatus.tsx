import { Divider } from '@mantine/core'
import { IconFile, IconFolder, IconX } from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { humanFileSize } from '@weblens/util'

import './style/uploadStatusStyle.scss'

import { useMemo } from 'react'
import '@weblens/components/style.scss'
import { SingleUpload, useUploadStatus } from './UploadStateControl'

function UploadCard({
    uploadMetadata,
    subUploads,
}: {
    uploadMetadata: SingleUpload
    subUploads: SingleUpload[]
}) {
    let prog = 0
    let statusText = ''
    let speed = 0
    if (uploadMetadata.isDir) {
        if (uploadMetadata.files === -1) {
            prog = -1
        } else {
            prog = (uploadMetadata.files / uploadMetadata.total) * 100
        }
        statusText = `${uploadMetadata.files} of ${uploadMetadata.total} files`
        speed = subUploads.reduce((acc, f) => f.getSpeed() + acc, 0)
    } else if (uploadMetadata.chunks.length !== 0) {
        const soFar = uploadMetadata.chunks.reduce(
            (acc, chunk) => acc + (chunk ? chunk.bytesSoFar : 0),
            0
        )

        prog = (soFar / uploadMetadata.total) * 100
        const [soFarString, soFarUnits] = humanFileSize(soFar)
        const [totalString, totalUnits] = humanFileSize(uploadMetadata.total)
        statusText = `${soFarString}${soFarUnits} of ${totalString}${totalUnits}`
        speed = uploadMetadata.getSpeed()
    }

    const [speedStr, speedUnits] = humanFileSize(speed)

    return (
        <div className="flex w-full flex-col p-2">
            <div className="flex flex-row h-max min-h-[40px] shrink-0 m-[1px] items-center">
                <div className="flex flex-col h-max w-0 items-start justify-center grow">
                    <p className="truncate font-semibold text-white w-full">
                        {uploadMetadata.friendlyName}
                    </p>
                    {statusText && prog !== 100 && prog !== -1 && (
                        <div>
                            <p className="text-nowrap pr-[4px] text-sm my-1">
                                {statusText}
                            </p>
                            {!uploadMetadata.isDir && (
                                <p className="text-nowrap pr-[4px] text-sm mt-1">
                                    {speedStr} {speedUnits}/s
                                </p>
                            )}
                        </div>
                    )}
                </div>
                {uploadMetadata.isDir && (
                    <IconFolder
                        color="white"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
                {!uploadMetadata.isDir && (
                    <IconFile
                        color="white"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
            </div>

            {!uploadMetadata.complete && (
                <WeblensProgress
                    value={prog}
                    failure={Boolean(uploadMetadata.error)}
                />
            )}
        </div>
    )
}

const UploadStatus = () => {
    const uploadsMap = useUploadStatus((state) => state.uploads)
    const clearUploads = useUploadStatus((state) => state.clearUploads)

    const uploadCards = useMemo(() => {
        const uploadCards = []

        const parents: SingleUpload[] = []
        const childrenMap = new Map<string, SingleUpload[]>()
        for (const upload of Array.from(uploadsMap.values())) {
            if (upload.parent) {
                if (childrenMap.get(upload.parent)) {
                    childrenMap.get(upload.parent).push(upload)
                } else {
                    childrenMap.set(upload.parent, [upload])
                }
            } else {
                parents.push(upload)
            }
        }
        parents.sort((a, b) => {
            const aVal = a.bytes / a.total
            const bVal = b.bytes / b.total
            if (aVal === bVal) {
                return 0
            } else if (aVal !== 1 && bVal === 1) {
                return -1
            } else if (bVal !== 1 && aVal === 1) {
                return 1
            } else if (aVal >= 0 && aVal <= 1) {
                return 1
            }

            return 0
        })

        for (const uploadMeta of parents) {
            uploadCards.push(
                <UploadCard
                    key={uploadMeta.key}
                    uploadMetadata={uploadMeta}
                    subUploads={childrenMap.get(uploadMeta.key)}
                />
            )
        }
        return uploadCards
    }, [uploadsMap])

    if (uploadCards.length === 0) {
        return null
    }

    const topLevelCount = Array.from(uploadsMap.values()).filter(
        (val) => !val.parent
    ).length

    return (
        <div className="upload-status-container">
            <div className="flex flex-col h-max max-h-full w-full bg-[#ffffff11] p-2 pb-0 mb-1 rounded overflow-hidden">
                <div className="flex no-scrollbar h-max min-h-[50px]">
                    <div className="h-max min-h-max w-full">{uploadCards}</div>
                </div>

                <Divider h={2} w={'100%'} />
                <div className="flex flex-row justify-center w-full h-max p-2">
                    <div className="flex flex-row h-full w-full items-center justify-between">
                        <p className="text-white font-semibold text-lg">
                            Uploading {topLevelCount} item
                            {topLevelCount !== 1 ? 's' : ''}
                        </p>
                        <WeblensButton
                            Left={IconX}
                            squareSize={30}
                            onClick={clearUploads}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export default UploadStatus
