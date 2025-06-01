import { Dimensions } from '@weblens/types/Types'
import { RefObject, useCallback, useEffect, useState } from 'react'

export function useResize(
    elem: RefObject<HTMLElement | null>,
    resizeCallback?: (oldSize: Dimensions, newSize: Dimensions) => void
) {
    const [size, setSize] = useState({ height: -1, width: -1 })

    const handler = useCallback(() => {
        if (!elem.current) {
            return -1
        }

        const newSize = {
            height: elem.current.clientHeight,
            width: elem.current.clientWidth,
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
        if (elem && elem.current) {
            setSize({
                height: elem.current.clientHeight,
                width: elem.current.clientWidth,
            })
            const obs = new ResizeObserver(handler)
            obs.observe(elem.current)
            return () => obs.disconnect()
        } else if (!elem) {
            console.warn('useResize called with null element')
        }
    }, [handler, elem])

    return size
}
