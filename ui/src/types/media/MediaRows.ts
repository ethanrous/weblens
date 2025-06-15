import { ReactElement } from 'react'

import WeblensMedia from './Media'

export type GalleryRowItem = { m: WeblensMedia; w: number }

export type GalleryRowInfo = {
    rowHeight: number
    rowWidth: number
    items: GalleryRowItem[]
    element?: ReactElement
}

export function isGalleryRow(input?: GalleryRowInfo): input is GalleryRowInfo {
    if (input === undefined) {
        return false
    }

    return (
        input.rowHeight !== undefined &&
        input.rowWidth !== undefined &&
        input.items !== undefined
    )
}

export function GetMediaRows(
    medias: WeblensMedia[],
    rowHeight: number,
    viewWidth: number,
    marginSize: number
) {
    if (medias.length === 0 || viewWidth === -1) {
        return []
    }

    medias = [...medias]

    const ROW_WIDTH = viewWidth - 32

    // const sortDirection = 1
    // medias.sort((m1, m2) => {
    //     const val =
    //         (m2.GetCreateTimestampUnix() - m1.GetCreateTimestampUnix()) *
    //         sortDirection
    //     return val
    // })

    medias.forEach((m, i) => {
        if (i !== 0) {
            m.SetPrevLink(medias[i - 1])
        } else {
            m.SetPrevLink()
        }
        if (i !== medias.length - 1) {
            m.SetNextLink(medias[i + 1])
        } else {
            m.SetNextLink()
        }
    })

    const rows: GalleryRowInfo[] = []
    let currentRowWidth = 0
    let currentRow: GalleryRowItem[] = []

    let absIndex = 0

    while (true) {
        if (medias.length === 0) {
            if (currentRow.length !== 0) {
                rows.push({
                    rowHeight: rowHeight,
                    rowWidth: ROW_WIDTH,
                    items: currentRow,
                })
            }
            break
        }

        const m = medias.shift()

        if (!m) {
            break
        }

        if (m.GetHeight() === 0) {
            console.error('Attempt to display media with 0 height:', m.Id())
            continue
        }

        m.SetAbsIndex(absIndex)
        absIndex++

        // Calculate width given height "imageBaseScale", keeping aspect ratio
        const newWidth =
            Math.floor((rowHeight / m.GetHeight()) * m.GetWidth()) + marginSize

        // If we are out of media, and the image does not overflow this row, add it and break
        if (medias.length === 0 && !(currentRowWidth + newWidth > ROW_WIDTH)) {
            currentRow.push({ m: m, w: newWidth })

            rows.push({
                rowHeight: rowHeight,
                rowWidth: ROW_WIDTH,
                items: currentRow,
            })
            break
        }

        // If the image will overflow the window
        else if (currentRowWidth + newWidth > ROW_WIDTH) {
            const leftover = ROW_WIDTH - currentRowWidth
            let consuming = false
            if (newWidth / 2 < leftover || currentRow.length === 0) {
                currentRow.push({ m: m, w: newWidth })
                currentRowWidth += newWidth
                consuming = true
            }
            const marginTotal = currentRow.length * marginSize
            const rowScale =
                ((ROW_WIDTH - marginTotal) / (currentRowWidth - marginTotal)) *
                rowHeight

            currentRow = currentRow.map((v) => {
                v.w = v.w * (rowScale / rowHeight)
                return v
            })

            rows.push({
                rowHeight: rowScale,
                rowWidth: ROW_WIDTH,
                items: currentRow,
            })
            currentRow = []
            currentRowWidth = 0

            if (consuming) {
                continue
            }
        }
        currentRow.push({ m: m, w: newWidth })
        currentRowWidth += newWidth
    }
    rows.unshift({ rowHeight: 20, rowWidth: ROW_WIDTH, items: [] })
    rows.push({ rowHeight: 20, rowWidth: ROW_WIDTH, items: [] })
    return rows
}
