export function capitalizeFirstLetter(str: string): string {
    if (!str) return str
    return str.charAt(0).toUpperCase() + str.slice(1)
}

export function camelCaseToWords(str: string): string {
    if (!str) return str
    const spaced = str.replace(/([a-z])([A-Z])/g, '$1 $2')
    return capitalizeFirstLetter(spaced)
}
