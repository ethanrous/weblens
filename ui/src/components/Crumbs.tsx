import { useNavigate } from 'react-router-dom'

import { Box, Text, Tooltip } from '@mantine/core'
import { IconChevronRight } from '@tabler/icons-react'

import { memo, useContext, useMemo, useState } from 'react'
import { fileData, getBlankFile } from '../types/Types'
import { userContext } from '../Context'
import { RowBox } from '../Pages/FileBrowser/FilebrowserStyles'

type breadcrumbProps = {
    label: string
    onClick?: React.MouseEventHandler<HTMLDivElement>
    onMouseUp?: () => void
    tooltipText?: string
    doCopy?: boolean
    dragging?: number
    sx?: any
    fontSize?: number
    alwaysOn?: boolean
}

export const StyledBreadcrumb = ({ label, onClick, tooltipText, doCopy, dragging, onMouseUp, sx, alwaysOn = false, fontSize = 25 }: breadcrumbProps) => {
    const [success, setSuccess] = useState(false)
    const [hovering, setHovering] = useState(false)

    if (doCopy) {
        tooltipText = `Copy "${label}"`
        onClick = (e) => {
            e.stopPropagation()
            navigator.clipboard.writeText(label)
            setSuccess(true)
            setTimeout(() => setSuccess(false), 1000)
        }
    } else { tooltipText = `${dragging === 1 ? "Move to" : "Go to"} ${label}` }
    let outline
    let bgColor
    if (success) {
        bgColor = "rgba(5, 125, 5, 1.0)"
        outline = '1px solid #aaaaaa'
    }
    else if (alwaysOn) {
        outline = '1px solid #aaaaaa'
        bgColor = "rgba(30, 30, 30, 0.5)"
    }
    else if (dragging === 1 && hovering) {
        outline = '1px solid #ffffff'
        bgColor = "rgb(30, 30, 90)"
    } else if (dragging === 1) {
        outline = '1px solid #aaaaaa'
        bgColor = "transparent"
    } else {
        outline = ""
        bgColor = "transparent"
    }
    return (
        <Tooltip label={tooltipText} >
            <Box
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                onMouseUp={onMouseUp}
                onClick={onClick}

                style={{ height: 'max-content', cursor: 'pointer', outline: outline, borderRadius: 4, backgroundColor: bgColor, padding: 2, margin: 8 }}
            >
                <Text lineClamp={1} c={'white'} truncate='end' style={{ fontSize: `${fontSize}px`, lineHeight: "1.2", userSelect: "none", width: '100%' }}>{label}</Text>
            </Box>
        </Tooltip >
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

const StyledLoaf = ({ crumbs }) => {
    return (
        <RowBox>
            {crumbs.map((c, i) => <Crumbcatenator key={i} crumb={c} last={i === crumbs.length - 1} />)}
        </RowBox>
    )
}

const Crumbs = memo(({ finalFile, parents, moveSelectedTo, navOnLast, dragging }: { finalFile: fileData, parents: fileData[], navOnLast: boolean, moveSelectedTo?: (folderId: string) => void, dragging?: number }) => {
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
            return <StyledBreadcrumb key={parent.id} label={isHome ? "Home" : parent.filename} onClick={(e) => { e.stopPropagation(); navigate(`/files/${isHome ? "home" : parent.id}`) }} doCopy={false} dragging={dragging} onMouseUp={() => { if (dragging !== 0) { moveSelectedTo(parent.id) } }} />
        })

        crumbs.push(
            <StyledBreadcrumb key={finalFile.id} label={finalFile.id === userInfo.homeFolderId ? "Home" : finalFile.filename} onClick={(e) => { e.stopPropagation(); if (!navOnLast) { return }; navigate(`/files/${finalFile.parentFolderId === userInfo.homeFolderId ? "home" : finalFile.parentFolderId}`) }} doCopy={!navOnLast} />
        )

        return (
            <StyledLoaf crumbs={crumbs} />
            // <StyledLoaf crumbs={crumbs} separator={" › "} />
        )
    }, [parents, finalFile, moveSelectedTo, navOnLast, dragging, navigate, userInfo])

    return loaf

}, (prev, next) => {
    return (prev.dragging === next.dragging && prev.parents === next.parents && prev.finalFile === next.finalFile) || next.parents === null || !next.finalFile?.id
})

export default Crumbs