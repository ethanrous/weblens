import { Dimensions } from '@weblens/types/Types'
import {
    RefObject,
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react'

export const useResize = (
    elem: HTMLDivElement,
    resizeCallback?: (oldSize: Dimensions, newSize: Dimensions) => void
) => {
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

export const useVideo = (elem: HTMLVideoElement) => {
    const [playtime, setPlaytime] = useState(0)
    const [isPlaying, setIsPlaying] = useState(false)
    const [isWaiting, setIsWaiting] = useState(true)

    const updatePlaytime = useCallback(() => {
        setPlaytime(elem.currentTime)
    }, [setPlaytime, elem])

    const updatePlayState = useCallback(
        (e: Event) => {
            setIsPlaying(e.type === 'play')
        },
        [setIsPlaying]
    )

    const updateBufferState = useCallback(
        (e: Event) => {
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
        (event: KeyboardEvent) => {
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

export const useClick = (
    handler: (e: MouseEvent) => void,
    ignore?: HTMLDivElement,
    disable?: boolean
) => {
    const callback = useCallback(
        (e: MouseEvent) => {
            if (disable) {
                return
            }

            if (ignore && ignore.contains(e.target as Node)) {
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

export const useTimer = (startTime: Date, startPaused?: boolean) => {
    const [elapsedTime, setElapsedTime] = useState(
        startTime ? Date.now() - startTime.getTime() : 0
    )
    const [isRunning, setIsRunning] = useState(false)
    const countRef = useRef<NodeJS.Timeout>(null)

    const handleStart = () => {
        if (isRunning) {
            return
        }
        setIsRunning(true)
        setElapsedTime(0)
        const startNum = startTime ? startTime.getTime() : 0
        countRef.current = setInterval(() => {
            setElapsedTime(Date.now() - startNum)
        }, 27)
    }

    useEffect(() => {
        clearInterval(countRef.current)
        if (!startTime || startPaused == false) {
            return
        }
        handleStart()
    }, [startTime])

    const handlePause = () => {
        clearInterval(countRef.current)
        setIsRunning(false)
    }

    const handleReset = () => {
        clearInterval(countRef.current)
        setIsRunning(false)
        setElapsedTime(0)
    }

    return { elapsedTime, isRunning, handleStart, handlePause, handleReset }
}
