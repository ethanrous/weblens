import { useEffect, useState } from 'react'

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
