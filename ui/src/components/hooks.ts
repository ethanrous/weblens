import { RefObject, useCallback, useEffect, useMemo, useState } from 'react'
import { mediaType } from '../types/Types'
import { fetchMediaTypes } from '../Media/MediaQuery'

export const useResize = (
    elem: HTMLDivElement,
    resizeCallback?: (oldSize, newSize) => void
) => {
    const [size, setSize] = useState({ height: -1, width: -1 })

    const handler = useCallback(() => {
        const newSize = {
            height: elem.clientHeight,
            width: elem.clientWidth,
        }
        // only 1 entry
        setSize((prev) => {
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

export const useVideo = (elem: HTMLVideoElement) => {
    const [playtime, setPlaytime] = useState(0)
    const [isPlaying, setIsPlaying] = useState(false)
    const [isWaiting, setIsWaiting] = useState(true)

    const updatePlaytime = useCallback(() => {
        setPlaytime(elem.currentTime)
    }, [setPlaytime, elem])

    const updatePlayState = useCallback(
        (e) => {
            setIsPlaying(e.type === 'play')
        },
        [setIsPlaying]
    )

    const updateBufferState = useCallback(
        (e) => {
            if (e.type === 'waiting') {
                setIsWaiting(true)
            } else if (e.type === 'playing') {
                setIsWaiting(false)
            }
        },
        [setIsWaiting]
    )

    useEffect(() => {
        if (!elem) {
            return
        }
        elem.addEventListener('timeupdate', updatePlaytime)
        elem.addEventListener('play', updatePlayState)
        elem.addEventListener('pause', updatePlayState)
        elem.addEventListener('waiting', updateBufferState)
        elem.addEventListener('playing', updateBufferState)
        return () => {
            elem.removeEventListener('timeupdate', updatePlaytime)
            elem.removeEventListener('play', updatePlayState)
            elem.removeEventListener('pause', updatePlayState)
            elem.removeEventListener('waiting', updateBufferState)
            elem.removeEventListener('playing', updateBufferState)
        }
    }, [updatePlaytime, updatePlayState, elem])

    return { playtime, isPlaying, isWaiting }
}

export const useKeyDown = (
    key: string | ((e: KeyboardEvent) => boolean),
    callback: (e: KeyboardEvent) => void,
    disable?: boolean
) => {
    const onKeyDown = useCallback(
        (event) => {
            if (
                (typeof key === 'string' && event.key === key) ||
                (typeof key === 'function' && key(event))
            ) {
                callback(event)
            }
        },
        [key, callback]
    )

    useEffect(() => {
        if (disable === true) {
            return
        }
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [onKeyDown, disable])
}

export const useWindowSize = () => {
    const [windowSize, setWindowSize] = useState({
        width: window.innerWidth,
        height: window.innerHeight,
    })
    const onResize = () => {
        setWindowSize({
            width: window.innerWidth,
            height: window.innerHeight,
        })
    }

    useEffect(() => {
        window.addEventListener('resize', onResize)
        return () => window.removeEventListener('resize', onResize)
    }, [onResize])

    return windowSize
}

export const useWindowSizeAgain = (
    updateCallback?: (oldSize, newSize) => void
) => {
    const [windowSize, setWindowSize] = useState({
        width: window.innerWidth,
        height: window.innerHeight,
    })

    const onResize = useCallback(
        (e) => {
            if (updateCallback) {
                updateCallback(windowSize, {
                    width: e.target.innerWidth,
                    height: e.target.innerHeight,
                })
            }
            setWindowSize({
                width: e.target.innerWidth,
                height: e.target.innerHeight,
            })
        },
        [setWindowSize, updateCallback]
    )

    useEffect(() => {
        window.addEventListener('resize', onResize)
        return () => window.removeEventListener('resize', onResize)
    }, [onResize])

    return windowSize
}

export const useResizeDrag = (
    resizing: boolean,
    setResizing: (r: boolean) => void,
    setResizeOffset: (o: number) => void,
    flip?: boolean,
    vertical?: boolean
) => {
    const unDrag = useCallback(() => {
        setResizing(false)
    }, [setResizing])

    const moved = useCallback(
        (e) => {
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
        },
        [setResizeOffset]
    )

    useEffect(() => {
        if (resizing) {
            addEventListener('mousemove', moved)
            window.addEventListener('mouseup', unDrag)
        }
        return () => {
            removeEventListener('mousemove', moved)
            window.removeEventListener('mouseUp', unDrag)
        }
    }, [resizing, moved, unDrag])
}

export const useMediaType = (): Map<string, mediaType> => {
    const [typeMap, setTypeMap] = useState(null)

    useEffect(() => {
        const mediaTypes = new Map<string, mediaType>()
        fetchMediaTypes().then((mt) => {
            const mimes: string[] = Array.from(Object.keys(mt))
            for (const mime of mimes) {
                mediaTypes.set(mime, mt[mime])
            }
            setTypeMap(mediaTypes)
        })
    }, [])
    return typeMap
}

export const useClick = (handler: (e) => void, ignore?, disable?: boolean) => {
    const callback = useCallback(
        (e) => {
            if (disable) {
                return
            }

            if (ignore && ignore.contains(e.target)) {
                return
            }

            handler(e)
        },
        [handler, ignore, disable]
    )

    useEffect(() => {
        if (!disable) {
            window.addEventListener('click', callback, true)
            window.addEventListener('contextmenu', callback, true)
        } else {
            return
        }
        return () => {
            window.removeEventListener('click', callback, true)
            window.removeEventListener('contextmenu', callback, true)
        }
    }, [callback, disable])
}

export const useIsFocused = (element: HTMLDivElement) => {
    const [active, setActive] = useState<boolean>(undefined)

    const handleFocusIn = () => {
        setActive(true)
    }

    const handleFocusOut = () => {
        setActive(false)
    }

    useEffect(() => {
        if (!element) {
            return
        }
        if (document.activeElement === element) {
            setActive(true)
        }
        element.addEventListener('focusin', handleFocusIn)
        element.addEventListener('focusout', handleFocusOut)
        return () => {
            document.removeEventListener('focusin', handleFocusIn)
            document.removeEventListener('focusout', handleFocusOut)
        }
    }, [element])

    return active
}

export function useOnScreen(ref: RefObject<HTMLElement>) {
    const [isIntersecting, setIntersecting] = useState(false)

    const observer = useMemo(
        () =>
            new IntersectionObserver(([entry]) =>
                setIntersecting(entry.isIntersecting)
            ),
        [ref]
    )

    useEffect(() => {
        observer.observe(ref.current)
        return () => observer.disconnect()
    }, [])

    return isIntersecting
}
