import { Box, Tooltip, useTheme, styled, Breadcrumbs, Chip } from '@mui/joy'
import { SxProps, Theme } from '@mui/joy/styles/types'
import { Dispatch, useState } from 'react'

type breadcrumbProps = {
    label: string
    dispatch?: Dispatch<any>
    onClick?: React.MouseEventHandler<HTMLDivElement>
    tooltipText?: string
    doCopy?: boolean
    sx?: any
    path?: string
}

export const StyledBreadcrumb = ({ label, onClick, tooltipText, doCopy, sx, path, dispatch }: breadcrumbProps) => {
    const [success, setSuccess] = useState(false)
    const [hovering, setHovering] = useState(false)
    const theme = useTheme()

    let backgroundColor
    if (success) {
        backgroundColor = theme.colorSchemes.dark.palette.success.solidBg
    } else if (hovering) {
        backgroundColor = theme.colorSchemes.dark.palette.primary.solidBg
    } else {
        backgroundColor = theme.colorSchemes.dark.palette.primary.solidDisabledBg
    }

    if (doCopy) {
        tooltipText = "Copy"
        onClick = (e) => {
            e.stopPropagation()
            navigator.clipboard.writeText(label)
            setSuccess(true)
            setTimeout(() => setSuccess(false), 1000)
        }
    }
    return (
        <Tooltip title={success ? "Copied!" : tooltipText} disableInteractive>
            <Box
                height={"max-content"}
                flexShrink={1}
                minWidth={0}
                onMouseOver={() => setHovering(true)}
                onMouseLeave={() => setHovering(false)}
                onClick={onClick}
                onMouseUp={() => { if (dispatch) { dispatch({ type: 'move_selected', targetItemPath: path, ignoreMissingItem: true }) } }}
                sx={{ ...sx, cursor: "pointer" }}
            >
                <Chip
                    variant='solid'
                    sx={{
                        borderRadius: "5px",
                        width: "100%",
                        backgroundColor,
                        height: 25,
                        userSelect: 'none'
                    }}
                >
                    {label}
                </Chip>
            </Box>
        </Tooltip >
    )
}

const StyledLoaf = ({ ...props }) => {
    const theme = useTheme()
    return (
        <Breadcrumbs
            {...props}
            sx={{
                backgroundColor: theme.colorSchemes.dark.palette.primary.softBg,
                outline: theme.colorSchemes.dark.palette.primary.outlinedColor,
                width: "max-content",
                borderRadius: "5px",
                ".MuiBreadcrumbs-separator": {
                    color: theme.colorSchemes.dark.palette.text.primary
                },
                margin: "20px"
            }}
        />
    )
}

const Crumbs = ({ path, includeHome, navOnLast, navigate, dispatch }) => {
    if (!path) {
        return (null)
    }
    try {
        path = `/${path}/`.replace(/\/\/+/g, '/').slice(1, -1)
        let parts: string[] = path.split('/')
        if (parts[0] == '' && parts.length == 1) {
            parts = ['/']
        } else {
            parts.unshift('/')
        }

        if (!includeHome) {
            parts = parts.slice(1)
        }

        let crumbPaths = []
        for (const [index, val] of parts.entries()) {
            if (index === 0) {
                crumbPaths.push(val)
                continue
            }
            let subPath = crumbPaths[index - 1]
            if (subPath.slice(-1) !== "/") {
                subPath += '/'
            }
            subPath += val

            crumbPaths.push(subPath)
        }

        const crumbs = parts.map((part, i) => {
            let navPath = "/files"
            if (crumbPaths[i][0] !== "/") { navPath += "/" }
            navPath += crumbPaths[i]
            let onClick = () => navigate(navPath)

            return <StyledBreadcrumb key={part} label={part == "/" ? "Home" : part} path={navPath} dispatch={dispatch} onClick={onClick} doCopy={!navOnLast && i == crumbPaths.length - 1} />
        })

        return (
            <StyledLoaf separator={"›"} >
                {crumbs}
            </StyledLoaf>
        )
    }
    catch (err) {
        console.log(err)
        return (null)
    }
}

export default Crumbs