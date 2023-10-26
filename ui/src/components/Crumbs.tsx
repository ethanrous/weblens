import Chip from '@mui/material/Chip'
import { emphasize, styled } from '@mui/material/styles'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { Box, Tooltip, alpha, useTheme } from '@mui/material'
import { useState } from 'react'

type breadcrumbProps = {
    label: string,
    onClick?: React.MouseEventHandler<HTMLDivElement>,
    tooltipText?: string
    doCopy?: boolean
}

export const StyledBreadcrumb = ({ label, onClick, tooltipText, doCopy }: breadcrumbProps) => {
    const [success, setSuccess] = useState(false)
    const theme = useTheme()

    let backgroundColor
    if (success) {
        backgroundColor = theme.palette.success.main
    }
    else {
        backgroundColor = alpha(theme.palette.background.default, 0.40)
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
            <Box height={"max-content"} padding={"5px"} flexShrink={1} minWidth={0} onClick={onClick} sx={{ cursor: "pointer" }}>
                <Chip label={label} sx={{
                    borderRadius: "5px",
                    backdropFilter: "blur(8px)",
                    width: "100%",
                    backgroundColor,
                    height: 25,
                    color: 'secondary',
                    fontWeight: theme.typography.fontWeightBold,

                    '&:hover, &:focus': {
                        backgroundColor: emphasize(backgroundColor, 0.15),
                    },
                    '&:active': {
                        boxShadow: theme.shadows[1],
                        backgroundColor: emphasize(backgroundColor, 0.12),
                    },
                }}
                />
            </Box>
        </Tooltip >
    )
}

const StyledLoaf = styled(Breadcrumbs)(({ theme }) => ({
    backgroundColor: alpha(theme.palette.background.default, 0.40),
    backdropFilter: "blur(8px)",
    borderRadius: "5px",
    ".MuiBreadcrumbs-separator": {
        color: theme.palette.primary.main
    }
}))

const Crumbs = ({ path, includeHome, navOnLast, navigate }) => {
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
            let onClick
            if (!navOnLast && i == crumbPaths.length - 1) {
                onClick = "copy"
            } else if (`/${path}` !== crumbPaths[i]) {
                let navPath = "/files"
                if (crumbPaths[i][0] !== "/") { navPath += "/" }
                navPath += crumbPaths[i]
                onClick = () => navigate(navPath)
            }
            return <StyledBreadcrumb key={part} label={part == "/" ? "Home" : part} onClick={onClick} />
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