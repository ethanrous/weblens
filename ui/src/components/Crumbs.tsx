import { Box, useTheme, styled, Breadcrumbs, Chip } from '@mui/joy'
import { Dispatch, useContext, useState } from 'react'
import { itemData } from '../types/Types'
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
                <Text lineClamp={1} c={'white'} style={{ textOverflow: 'ellipsis', fontSize: `${fontSize}px`, lineHeight: "1", userSelect: "none" }}>{label}</Text>
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
                width: "max-content",
                borderRadius: "3px",
                ".MuiBreadcrumbs-separator": {
                    color: theme.colorSchemes.dark.palette.text.primary
                },
                // margin: "10px"
            }}
        />
    )
}

const Crumbs = ({ finalItem, parents, moveSelectedTo, navOnLast, dragging }: { finalItem: itemData, parents: itemData[], navOnLast: boolean, moveSelectedTo?: (folderId: string) => void, dragging?: number }) => {
    const navigate = useNavigate()
    const { userInfo } = useContext(userContext)
    if (parents === null || !finalItem?.id) {
        return null
    }
    if (parents.length != 0 && parents[0].id == "shared" && finalItem.id == "shared") {
        parents.shift()
    }
    if ((parents.length == 0 && finalItem.owner != userInfo.username && finalItem.id != "shared") || (parents.length != 0 && parents[0].id != "shared" && parents[parents.length - 1].owner != userInfo.username)) {
        let sharedRoot: itemData = {
            id: "shared",
            filename: "Shared",
            parentFolderId: "",
            owner: "",
            isDir: true,
            imported: false,
            modTime: new Date().toString(),
            size: 0,
            visible: true,
            mediaData: null
        }
        parents.unshift(sharedRoot)
    }

    try {
        const crumbs = parents.map((parent, i) => {
            const isHome = parent.id === userInfo.homeFolderId
            return <StyledBreadcrumb key={parent.id} label={isHome ? "Home" : parent.filename} onClick={(e) => { e.stopPropagation(); navigate(`/files/${isHome ? "home" : parent.id}`) }} doCopy={false} dragging={dragging} onMouseUp={() => { if (dragging !== 0) { moveSelectedTo(parent.id) } }} />
        })

        crumbs.push(
            <StyledBreadcrumb key={finalItem.id} label={finalItem.id === userInfo.homeFolderId ? "Home" : finalItem.filename} onClick={(e) => { e.stopPropagation(); if (!navOnLast) { return }; navigate(`/files/${finalItem.parentFolderId === userInfo.homeFolderId ? "home" : finalItem.parentFolderId}`) }} doCopy={!navOnLast} />
        )

        return (
            <StyledLoaf separator={" › "} >
                {crumbs}
            </StyledLoaf>
        )
    }
    catch (err) {
        console.error(err)
        return (null)
    }
}

export default Crumbs