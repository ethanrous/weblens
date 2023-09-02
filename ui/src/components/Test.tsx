import { Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import { useRef, useEffect, useReducer, useState } from 'react'

import InfiniteScroll from 'react-infinite-scroll-component';

const style = {
    //height: 30,
    border: "1px solid green",
    margin: 6,
    padding: 8
};

const Thing = ({ index }) => {
    return (
        <img style={style} key={index} src='http://localhost:3000/api/thumbnail?filehash=fAxjEOvXXnXJGfixz9VlS0T2eF7XgroKBqVxkF5IMUk%3D'>

        </img>
    );
};

const Test = () => {
    var [media, setMedia] = useState([])


    const fetchData = () => {
        console.log(media.length)
        setMedia(prevItems => [...prevItems, ...Array.from({ length: 1 })])
    }

    useEffect(() => {
        setMedia(Array.from({ length: 10 }))
    }, []);

    return (
        <div>
            <InfiniteScroll
                dataLength={media.length}
                next={fetchData}
                hasMore={true}
                loader={<h1 style={{ textAlign: "center" }}>Loading...</h1>}
                endMessage={<p>No more data to load.</p>}
            >

                {media.map((value, index) => (
                    <Thing index={index} key={index} />
                ))}

            </InfiniteScroll>
        </div>
    )
}

export default Test