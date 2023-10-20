import Chip from '@mui/material/Chip'
import { createTheme, emphasize, styled } from '@mui/material/styles'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { Box } from '@mui/material'
import { useState } from 'react'

const customTheme = createTheme({
    palette: {
        primary: {
            main: '#1976d2',
            contrastText: 'white',
        }
    },
    typography: {
        fontWeightRegular: "12px"
    },

})


export const StyledBreadcrumb = ({ label, onClick }) => {
    const [success, setSuccess] = useState(false)
    let backgroundColor
    if (success) {
        backgroundColor = "rgb(0, 255, 0)"
    }
    else {
        backgroundColor = "rgb(230, 230, 230)"
    }

    if (onClick == "copy") {
        onClick = (e) => {
            console.log(e)
            e.stopPropagation()
            navigator.clipboard.writeText(label)
            setSuccess(true)
            setTimeout(() => setSuccess(false), 2000)
        }
    }

    return (
        <Box height={"max-content"} width={"100%"} onClick={onClick} sx={{ cursor: "pointer" }}>
            <Chip label={success ? "Copied" : label} sx={{
                width: "100%",
                backgroundColor,
                height: 25,
                color: "white",
                fontWeight: customTheme.typography.fontWeightRegular,
                '&:hover, &:focus': {
                    backgroundColor: emphasize(backgroundColor, 0.10),
                },
                '&:active': {
                    boxShadow: customTheme.shadows[1],
                    backgroundColor: emphasize(backgroundColor, 0.12),
                },
            }}
            />

        </Box>
    )
}

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
            } else {
                let navPath = "/files"
                if (crumbPaths[i][0] !== "/") {
                    navPath += "/"
                }
                navPath += crumbPaths[i]
                onClick = () => navigate(navPath)
            }
            return <StyledBreadcrumb key={part} label={part == "/" ? "Home" : part} onClick={onClick} />
        })

        return (
            <Breadcrumbs separator={"›"} >
                {crumbs}
            </Breadcrumbs>
        )
    }
    catch (err) {
        console.log(err)
        return (null)
    }
}

export default Crumbs