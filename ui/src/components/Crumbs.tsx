import Chip from '@mui/material/Chip'
import { createTheme, emphasize, styled } from '@mui/material/styles'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { Box } from '@mui/material'

// export const StyledBreadcrumb = styled(Chip)(({ theme, success }) => {
//     let backgroundColor
//     if (success) {

//     }
//     else {
//         backgroundColor =
//             theme.palette.mode === 'light'
//                 ? theme.palette.grey[100]
//                 : theme.palette.grey[800]
//     }
//     return {
//         backgroundColor,
//         height: theme.spacing(3.5),
//         color: theme.palette.text.primary,
//         fontWeight: theme.typography.fontWeightRegular,
//         '&:hover, &:focus': {
//             backgroundColor: emphasize(backgroundColor, 0.06),
//         },
//         '&:active': {
//             boxShadow: theme.shadows[1],
//             backgroundColor: emphasize(backgroundColor, 0.12),
//         },
//     }
// }) as typeof Chip

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


export const StyledBreadcrumb = ({ label, success, onClick }) => {
    let backgroundColor
    if (success) {
        backgroundColor = "rgb(0, 255, 0)"
    }
    else {
        backgroundColor = "rgb(230, 230, 230)"
    }
    return (
        <Box height={"max-content"} width={"100%"} onClick={onClick} sx={{ cursor: "pointer" }}>
            <Chip label={label} sx={{
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

const Crumbs = ({ path, includeHome, navigate }) => {
    if (!path) {
        return (null)
    }
    try {

        path = path.slice(1)
        let parts = path.split('/')
        while (parts[parts.length - 1] == '') {
            parts.pop()
        }

        if (includeHome) {
            parts.unshift('/')
        }
        const current = parts.pop()

        let crumbPaths = []
        for (let [index, val] of parts.entries()) {
            if (index == 0 && includeHome) {
                crumbPaths.push("/")
                continue
            } else {
                crumbPaths.push(crumbPaths[index - 1] + "/" + val)
            }
        }
        const crumbs = parts.map((part, i) => (
            <StyledBreadcrumb key={part} label={part == "/" ? "Home" : part} success={false} onClick={() => { navigate(`/files/${crumbPaths[i]}`.replace(/\/\/+/g, '/')) }} />)
        )
        crumbs.push(
            <StyledBreadcrumb key={current} label={current == "/" ? "Home" : current} success={false} onClick={() => { }} />
        )
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