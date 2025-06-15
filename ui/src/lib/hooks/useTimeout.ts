import { useCallback, useEffect, useState } from 'react'

export const useTimeout = (
    onTimeout: () => void,
    delay: number,
    opts?: { onStart?: () => void }
) => {
    const [timeout, setTo] = useState<NodeJS.Timeout | null>(null)

    const clear = () => {
        if (timeout) {
            clearTimeout(timeout)
            setTo(null)
        }
    }

    const reset = useCallback(
        (delayOverride?: number) => {
            if (timeout) {
                clearTimeout(timeout)
            }

            const newTimeout: NodeJS.Timeout = setTimeout<any>(() => {
                console.log('TIMEOUT')
                onTimeout()
            }, delayOverride ?? delay)

            if (opts?.onStart) {
                opts.onStart()
            }

            setTo(newTimeout)
        },
        [onTimeout, timeout]
    )

    useEffect(() => {
        reset()
    }, [])

    return { clear, reset }
}
