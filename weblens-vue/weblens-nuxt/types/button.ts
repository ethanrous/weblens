export type ButtonProps = {
    label?: string
    errorText?: string | ((e: Error) => string)

    type?: 'default' | 'outline' | 'light'
    flavor?: 'primary' | 'danger' | 'secondary'
    danger?: boolean
    disabled?: boolean
    selected?: boolean

    squareSize?: number
    fillWidth?: boolean
    centerContent?: boolean
    allowCollapse?: boolean
    merge?: 'row' | 'column'

    onClick?: ((e: MouseEvent) => Promise<void>) | ((e: MouseEvent) => void)
}
