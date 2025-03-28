import { MouseEvent, useEffect, useRef, useState } from 'react'

export function useMouse<T extends HTMLElement>(
    options: { resetOnExit?: boolean } = { resetOnExit: false }
) {
    const [position, setPosition] = useState({ x: 0, y: 0 })

    const ref = useRef<T>(null)

    const setMousePosition = (event: MouseEvent<HTMLElement>) => {
        if (ref.current) {
            const rect = event.currentTarget.getBoundingClientRect()

            const x = Math.max(
                0,
                Math.round(event.pageX - rect.left - window.scrollX)
            )

            const y = Math.max(
                0,
                Math.round(event.pageY - rect.top - window.scrollY)
            )

            setPosition({ x, y })
        } else {
            setPosition({ x: event.clientX, y: event.clientY })
        }
    }

    const resetMousePosition = () => setPosition({ x: 0, y: 0 })

    useEffect(() => {
        const element = ref?.current ? ref.current : document
        element.addEventListener(
            'mousemove',
            setMousePosition as unknown as EventListenerOrEventListenerObject
        )
        if (options.resetOnExit) {
            element.addEventListener(
                'mouseleave',
                resetMousePosition as unknown as EventListenerOrEventListenerObject
            )
        }

        return () => {
            element.removeEventListener(
                'mousemove',
                setMousePosition as unknown as EventListenerOrEventListenerObject
            )
            if (options.resetOnExit) {
                element.removeEventListener(
                    'mouseleave',
                    resetMousePosition as unknown as EventListenerOrEventListenerObject
                )
            }
        }
    }, [ref.current])

    return { ref, ...position }
}
