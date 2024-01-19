import { Box, useTheme, Breadcrumbs } from '@mui/joy'
import { memo, useContext, useEffect, useMemo, useState } from 'react'
import { fileData, getBlankFile } from '../types/Types'
import { userContext } from '../Context'
import { useNavigate } from 'react-router-dom'
import { Text, Tooltip } from '@mantine/core'

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
    } else { tooltipText = `Go to ${label}` }
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
    else if (dragging && hovering) {
        outline = '1px solid #ffffff'
        bgColor = "rgb(30, 30, 90)"
    } else if (dragging) {
        outline = '1px solid #aaaaaa'
        bgColor = "transparent"
    } else {
        outline = ""
        bgColor = "transparent"
    }
    return (
        <Tooltip label={tooltipText} >
            <Box
                height={"max-content"}
                flexShrink={1}
                minWidth={0}
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                onMouseUp={onMouseUp}
                onClick={onClick}
                padding={1}
                sx={{ ...sx, cursor: "pointer", outline: outline, borderRadius: "5px", backgroundColor: bgColor }}
            >
                <Text lineClamp={1} c={'white'} truncate='end' style={{ fontSize: `${fontSize}px`, lineHeight: "1.2", userSelect: "none", width: '100%' }}>{label}</Text>
            </Box>
        </Tooltip >
    )
}

const StyledLoaf = ({ ...props }) => {
    const theme = useTheme()
    return (
        <Breadcrumbs
            {...props}
            size='lg'
            sx={{
                // width: "100%",
                borderRadius: "3px",
                ".MuiBreadcrumbs-separator": {
                    color: theme.colorSchemes.dark.palette.text.primary
                },
            }}
        />
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
            <StyledLoaf separator={" › "} >
                {crumbs}
            </StyledLoaf>
        )
    }, [parents, finalFile, moveSelectedTo, navOnLast, dragging, navigate, userInfo])

    return loaf

}, (prev, next) => {
    return (prev.dragging === next.dragging && prev.parents === next.parents && prev.finalFile === next.finalFile) || next.parents === null || !next.finalFile?.id
})

export default Crumbs