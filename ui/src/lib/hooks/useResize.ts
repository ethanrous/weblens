import { Dimensions } from '@weblens/types/Types'
import { useCallback, useEffect, useState } from 'react'

export function useResize(
    elem: HTMLElement,
    resizeCallback?: (oldSize: Dimensions, newSize: Dimensions) => void
) {
    const [size, setSize] = useState({ height: -1, width: -1 })

    const handler = useCallback(() => {
        const newSize = {
            height: elem.clientHeight,
            width: elem.clientWidth,
        }
        // only 1 entry
        setSize((prev: Dimensions) => {
            if (resizeCallback) {
                resizeCallback(prev, newSize)
            }
            return newSize
        })
    }, [elem, resizeCallback])

    useEffect(() => {
        if (elem) {
            setSize({ height: elem.clientHeight, width: elem.clientWidth })
            const obs = new ResizeObserver(handler)
            obs.observe(elem)
            return () => obs.disconnect()
        }
    }, [handler])

    return size
}
