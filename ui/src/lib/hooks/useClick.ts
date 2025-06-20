import { RefObject, useCallback, useEffect } from 'react'

export const useClick = (
    handler: (e: MouseEvent) => void,
    ignore?: RefObject<HTMLElement | null>,
    disable?: boolean
) => {
    const callback = useCallback(
        (e: MouseEvent) => {
            if (disable) {
                return
            }

            if (ignore?.current && ignore.current.contains(e.target as Node)) {
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
