import { useRef, useEffect, useReducer, useState } from 'react'
import InfiniteScroll from 'react-infinite-scroll-component';

import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import { useNavigate } from "react-router-dom";

import PhotoContainer from './PhotoContainer'
import styled from '@emotion/styled';

const DateWrapper = ({ dateTime }) => {
    let dateObj = new Date(dateTime)
    let dateStr = dateObj.toUTCString().split(" 00:00:00 GMT")[0]
    return (
        <Box
            key={`${dateStr} title`}
            component="h3"
            fontSize={35}
            pl={2}
        >
            {dateStr}
        </Box>
    )
}

const BlankCard = styled("div")({
    height: '250px',
    flexGrow: 999999
})


const BucketCards = ({ medias }) => {
    let mediaCards = medias.map((mediaData) => (
        <PhotoContainer
            key={mediaData.FileHash}
            mediaData={mediaData}
        />
    ))

    return (
        <Grid container
            display="flex"
            flexDirection="row"
            flexWrap="wrap"
            justifyContent="flex-start"
        >
            {mediaCards}
            <BlankCard />
        </Grid>
    )
}

const GalleryBucket = ({
    date,
    medias,
}) => {

    return (
        <Grid item >
            <DateWrapper dateTime={date} />
            <BucketCards medias={medias[date]} />
        </Grid >

    )
}

const mediaReducer = (state, action) => {
    switch (action.type) {
        case 'add_media': {
            return {
                buckets: action.buckets,
                total_items: action.total_items
            };
        }
    }
}

const fetchData = async (mediaState, setIsLoading, setError) => {
    setIsLoading(true);
    setError(null);

    try {
        var url = new URL("http:localhost:3000/api/media");
        url.searchParams.append('limit', '100')
        url.searchParams.append('skip', mediaState.total_items.toString())
        const response = await fetch(url.toString());
        const data = await response.json();

        let mediaBuckets = { ...mediaState.buckets }

        for (var item of data) {

            var [date, _] = item.CreateDate.split("T")
            if (mediaBuckets[date] == null) {
                mediaBuckets[date] = [item]
            } else {
                mediaBuckets[date].push(item)
            }
        }

        setIsLoading(false);
        return mediaBuckets

    } catch (error) {
        console.log("BAD VEFR BAD")
        console.log(error)
        setError(error);
    } finally {
        setIsLoading(false);
    }
};

const Gallery = () => {

    const [mediaState, dispatch] = useReducer(mediaReducer, {
        buckets: {},
        total_items: 0
    })

    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState(null)



    let navigate = useNavigate();
    const routeChange = () => {
        let path = `/upload`;
        navigate(path);
    }

    const moar_data = () => {
        fetchData(mediaState, setIsLoading, setError).then((data) => dispatch({
            type: 'add_media',
            buckets: data,
            total_items: mediaState.total_items + 100
        }))
    }

    useEffect(() => moar_data(), [])


    if ((!mediaState?.buckets || mediaState.buckets.size == 0) && !isLoading) {
        return (
            <Box
                display="flex"
                flexWrap="wrap"
                flexDirection="column"
                pt="50px"
                alignContent="center"
                gap="25px"
            >
                {"No media to display"}
                <Button onClick={routeChange}>
                    Upload Media
                </Button>
            </Box>
        )
    }

    const dateGroups = Object.keys(mediaState.buckets).map((value, i) => (
        <GalleryBucket date={value} medias={mediaState.buckets} key={value} />
    ))

    return (
        <Container maxWidth='xl'>
            <InfiniteScroll
                dataLength={mediaState.total_items}
                next={moar_data}
                children={dateGroups}
                hasMore={true}
                loader={<h1 style={{ textAlign: "center" }}>Loading...</h1>}
                endMessage={<p>No more data to load.</p>}
            />


        </Container>
    )
}

export default Gallery
