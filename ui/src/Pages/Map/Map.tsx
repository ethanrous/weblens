
import { useEffect, useRef, useState } from "react"
import { FlexColumnBox } from "../FileBrowser/FilebrowserStyles";

function MapPage() {
    const mapContainer = useRef(null);
    const [map, setMap] = useState(null);
    const [lng, setLng] = useState(-70.9);
    const [lat, setLat] = useState(42.35);
    const [zoom, setZoom] = useState(9);

    useEffect(() => {
        // setMap(new mapboxgl.Map({
        container: mapContainer.current,
            style: 'mapbox://styles/mapbox/streets-v12',
                center: [lng, lat],
                    zoom: zoom
    }))
}, [])
return (
    <FlexColumnBox>
        {map}
    </FlexColumnBox>
)
}

export default MapPage