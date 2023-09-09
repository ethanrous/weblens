import Gallery from "./Gallery"
import HeaderBar from "./HeaderBar"
import { Box } from '@mui/material';

const HomePage = () => {
    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: 'column',
            }}
        >

            <Gallery />
        </Box >
    )
}

export default HomePage