import { useRef, useEffect, useReducer, useState, useMemo } from 'react'
import Gallery from "./Gallery"
import HeaderBar from "./HeaderBar"
import { Box } from '@mui/material';

type GalleryOptions = {
    showRaw: boolean,

}

const defaultOpts: GalleryOptions = {
    showRaw: true
}

const HomePage = () => {

    const [galleryOpts, setGalOpts] = useState(defaultOpts)

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