import './slider.scss'
import { memo, useEffect, useMemo, useState } from 'react'
import { clamp } from '../util'

const WeblensSlider = memo(
    ({
        value,
        width,
        height,
        min,
        max,
        callback,
    }: {
        value: number
        width: number
        height?: number
        min: number
        max: number
        callback: (percent: number) => void
    }) => {
        const [dragging, setDragging] = useState(false)
        const [containerRef, setContainerRef] = useState(null)
        const [percentage, setPercentage] = useState(
            (value - min) / (max - min)
        )
        // const [bounceP] = useDebouncedValue(percentage, 50);
        useEffect(() => {
            callback(percentage * (max - min) + min)
        }, [percentage])
        // useEffect(() => {
        //     callback(bounceP * (max - min) + min);
        // }, [bounceP]);

        const left = useMemo(() => {
            if (containerRef) {
                return containerRef.getBoundingClientRect().left
            }
        }, [containerRef])

        useEffect(() => {
            if (dragging) {
                const update = (e) =>
                    setPercentage(clamp((e.clientX - left - 5) / width, 0, 1))
                // setPercentage(((e.clientX - left) / width) * (max - min) + min);
                // callback(;
                const stop = () => setDragging(false)
                window.addEventListener('mousemove', update)
                window.addEventListener('mouseup', stop)
                return () => {
                    window.removeEventListener('mousemove', update)
                    window.removeEventListener('mouseup', stop)
                }
            }
        }, [dragging])

        return (
            <div
                ref={setContainerRef}
                className="slider-container"
                style={{ width: width, height: height }}
                onMouseUp={() => setDragging(false)}
                onMouseDown={(e) => {
                    setPercentage(clamp((e.clientX - left - 5) / width, 0, 1))
                    setDragging(true)
                }}
            >
                <div
                    className="slider-handle"
                    style={{ left: percentage * width }}
                    onMouseDown={() => setDragging(true)}
                    onMouseUp={() => setDragging(false)}
                />
                <div
                    className="fill-bar"
                    style={{ width: percentage * width + 10 }}
                />
            </div>
        )
    },
    (prev, next) => {
        return true
    }
)

export default WeblensSlider
