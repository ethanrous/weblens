import Gallery from "./Gallery"
import HeaderBar from "./HeaderBar"
import Box from '@mui/material/Box';

const HomePage = () => {
    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: 'column',
            }}
        >
            <HeaderBar />
            <Gallery />
        </Box >
    )
}

export default HomePage