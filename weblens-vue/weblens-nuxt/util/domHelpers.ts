export function isParent(parent: Element, child: Element): boolean {
    if (!parent || !child) return false
    let currentElement = child.parentElement
    while (currentElement) {
        if (currentElement === parent) {
            return true
        }
        currentElement = currentElement.parentElement
    }
    return false
}

export function toCssUnit(value: number | string, defaultUnit = 'px'): string {
    if (typeof value === 'number') {
        return `${value}${defaultUnit}`
    }

    // If it's already a string, we check if it already has a unit
    if (typeof value === 'string') {
        // If it ends with 'px', 'em', '%', or 'rem', we return
        if (/^(?:\d+(\.\d+)?(px|em|%|rem))$/.test(value)) {
            return value
        }
        // If it doesn't have a unit, we assume it's a number and append the default unit
        return `${value}${defaultUnit}`
    }

    return value
}
