import { useNavigate } from 'react-router-dom'

import { Box, Text } from '@mantine/core'
import { IconChevronRight } from '@tabler/icons-react'

import { RowBox, TransferCard } from '../Pages/FileBrowser/FilebrowserStyles'
import { memo, useContext, useMemo, useState } from 'react'
import { fileData, getBlankFile } from '../types/Types'
import { userContext } from '../Context'

type breadcrumbProps = {
    label: string
    onClick?: React.MouseEventHandler<HTMLDivElement>
    onMouseUp?: () => void
    dragging?: number
    fontSize?: number
    compact?: boolean
    alwaysOn?: boolean
    setMoveDest?
}

export const StyledBreadcrumb = ({ label, onClick, dragging, onMouseUp, alwaysOn = false, fontSize = 25, compact = false, setMoveDest }: breadcrumbProps) => {
    const [hovering, setHovering] = useState(false)
    let outline
    let bgColor = "transparent"

    if (alwaysOn) {
        outline = '1px solid #aaaaaa'
        bgColor = "rgba(30, 30, 30, 0.5)"
    } else if (dragging === 1 && hovering) {
        outline = "2px solid #661199"
    } else if (dragging === 1) {
        bgColor = "#333333"
    }
    return (

        <Box
            className={compact ? 'crumb-box-compact' : 'crumb-box'}
            onMouseOver={() => { setHovering(true); if (dragging && setMoveDest) { setMoveDest(label) } }}
            onMouseLeave={() => { setHovering(false); if (dragging && setMoveDest) { setMoveDest("") } }}
            onMouseUp={e => { onMouseUp(); setMoveDest("") }}
            onClick={onClick}

            style={{ outline: outline, backgroundColor: bgColor }}
        >
            <Text lineClamp={1} c={'white'} truncate='end' style={{ fontSize: `${fontSize}px`, lineHeight: "1.2", userSelect: "none", width: '100%' }}>{label}</Text>
        </Box>

    )
}

// The crumb concatenator, the Crumbcatenator
const Crumbcatenator = ({ crumb, last }) => {
    return (
        <RowBox style={{ width: 'max-content' }}>
            {crumb}
            {!last && (
                <IconChevronRight />
            )}
        </RowBox>
    )
}

export const StyledLoaf = ({ crumbs }) => {
    return (
        <RowBox style={{ height: 'max-content' }}>
            {crumbs.map((c, i) => <Crumbcatenator key={i} crumb={c} last={i === crumbs.length - 1} />)}
        </RowBox>
    )
}

const Crumbs = memo(({ finalFile, parents, moveSelectedTo, navOnLast, setMoveDest, dragging }: { finalFile: fileData, parents: fileData[], navOnLast: boolean, moveSelectedTo?: (folderId: string) => void, setMoveDest?: (itemName: string) => void, dragging?: number }) => {
    const navigate = useNavigate()
    const { userInfo } = useContext(userContext)

    const loaf = useMemo(() => {
        if (!userInfo || !finalFile?.id) {
            return null
        }

        const parentsIds = parents.map(p => p.id)
        if (parentsIds.includes("shared")) {
            let sharedRoot = getBlankFile()
            sharedRoot.filename = "Shared"
            parents.unshift(sharedRoot)
        } else if (finalFile.id === userInfo.trashFolderId || parentsIds.includes(userInfo.trashFolderId)) {
            if (parents[0]?.id === userInfo.homeFolderId) {
                parents.shift()
            }
            if (parents[0]?.id === userInfo.trashFolderId && parents[0].filename !== "Trash") {
                parents[0].filename = "Trash"
            }
            if (finalFile.id === userInfo.trashFolderId && finalFile.filename !== "Trash") {
                finalFile.filename = "Trash"
            }
        }

        const crumbs = parents.map((parent) => {
            const isHome = parent.id === userInfo.homeFolderId
            return <StyledBreadcrumb key={parent.id} label={isHome ? "Home" : parent.filename} onClick={(e) => { e.stopPropagation(); navigate(`/files/${isHome ? "home" : parent.id}`) }} dragging={dragging} onMouseUp={() => { if (dragging !== 0) { moveSelectedTo(parent.id) } }} setMoveDest={setMoveDest} />
        })

        crumbs.push(
            <StyledBreadcrumb key={finalFile.id} label={finalFile.id === userInfo.homeFolderId ? "Home" : finalFile.filename} onClick={(e) => { e.stopPropagation(); if (!navOnLast) { return }; navigate(`/files/${finalFile.parentFolderId === userInfo.homeFolderId ? "home" : finalFile.parentFolderId}`) }} setMoveDest={setMoveDest} />
        )

        return (
            <StyledLoaf crumbs={crumbs} />
        )
    }, [parents, finalFile, moveSelectedTo, navOnLast, dragging, navigate, userInfo])

    return loaf

}, (prev, next) => {
    return (prev.dragging === next.dragging && prev.parents === next.parents && prev.finalFile === next.finalFile) || next.parents === null || !next.finalFile?.id
})

export default Crumbs