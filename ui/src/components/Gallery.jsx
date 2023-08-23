import * as React from 'react';

import Grid from '@mui/material/Grid';

import Box from '@mui/material/Box';
import Container from '@mui/material/Container';

import PhotoContainer from './PhotoContainer'

async function getImages() {
    const resp = await fetch("/api/photos")
    const body = await resp.json()
    var dates = {}
    for (var image in body) {
        var dateString = new Date(Date.parse(body[image].CreateDate)).toDateString()
        if (!dates[dateString]) {
            dates[dateString] = []
        }
        dates[dateString].push(body[image])
    }
    return dates
}

const start = Date.now();
const images = await getImages()
const end = Date.now();
console.log(`Fetch and parse images time: ${end - start} ms`);

const Gallery = () => {
    return (
        <Container maxWidth={false}>
            {Object.keys(images).map((date) => (
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        color: 'text.white',
                        padding: "30px"
                    }}>
                    <Box
                        component="span"
                        sx={{
                            fontSize: 34,
                            pb: '45px',
                        }}>
                        {date}
                    </Box>
                    <Grid container sx={{ flexGrow: 0.5 }}>
                        {images[date].map((image) => (
                            <PhotoContainer imageData={image}></PhotoContainer>
                        ))}
                    </Grid>

                </Box>
            ))}
        </Container >
    )
}

export default Gallery
