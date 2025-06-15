import { Dimensions } from '@weblens/types/Types'
import { RefObject, useCallback, useEffect, useState } from 'react'

type resizeOptions = {
    resizeCallback?: (oldSize: Dimensions, newSize: Dimensions) => void
    heightOffset?: number
    widthOffset?: number
}

export function useResize(
    elem: RefObject<HTMLElement | null>,
    opts?: resizeOptions
) {
    const [size, setSize] = useState({ height: -1, width: -1 })

    const handler = useCallback(() => {
        if (!elem.current) {
            return -1
        }

        const newSize = {
            height: elem.current.clientHeight + (opts?.heightOffset || 0),
            width: elem.current.clientWidth + (opts?.widthOffset || 0),
        }
        // only 1 entry
        setSize((prev: Dimensions) => {
            if (opts?.resizeCallback) {
                opts.resizeCallback(prev, newSize)
            }
            return newSize
        })
    }, [elem])

    useEffect(() => {
        if (elem && elem.current) {
            setSize({
                height: elem.current.clientHeight + (opts?.heightOffset || 0),
                width: elem.current.clientWidth + (opts?.widthOffset || 0),
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

export function useResizeWindow() {
    const [size, setSize] = useState({
        height: window.innerHeight,
        width: window.innerWidth,
    })

    const handler = useCallback(() => {
        setSize({
            height: window.innerHeight,
            width: window.innerWidth,
        })
    }, [])

    useEffect(() => {
        window.addEventListener('resize', handler)
        return () => window.removeEventListener('resize', handler)
    }, [handler])

    return size
}
