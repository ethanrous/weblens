import { useCallback, useEffect } from 'react'

export function useKeyDown(
    key: string | ((e: KeyboardEvent) => boolean),
    callback: (e: KeyboardEvent) => void,
    disable?: boolean
) {
    const onKeyDown = useCallback(
        (event: KeyboardEvent) => {
            if (
                !event.ctrlKey &&
                !event.metaKey &&
                ((typeof key === 'string' && event.key === key) ||
                    (typeof key === 'function' && key(event)))
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
