import { useEffect } from 'react'

export const useResizeDrag = (
    resizing: boolean,
    setResizing: (r: boolean) => void,
    setResizeOffset: (o: number) => void,
    flip?: boolean,
    vertical?: boolean
) => {
    const unDrag = (e: MouseEvent) => {
        e.stopPropagation()
        setResizing(false)
    }

    const moved = (e: MouseEvent) => {
        let val: number
        let windowSize: number
        if (vertical) {
            val = e.clientY
            windowSize = window.innerHeight
        } else {
            val = e.clientX
            windowSize = window.innerWidth
        }

        if (flip) {
            setResizeOffset(windowSize - val)
        } else {
            setResizeOffset(val)
        }
    }

    useEffect(() => {
        if (resizing) {
            window.addEventListener('mousemove', moved)
            window.addEventListener('mouseup', unDrag)
        }
        return () => {
            window.removeEventListener('mousemove', moved)
            window.removeEventListener('mouseup', unDrag)
        }
    }, [resizing, moved, unDrag])
}
