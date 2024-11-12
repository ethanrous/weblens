import { CSSProperties, FC } from 'react'

export type ButtonIcon = FC<{ className: string }>

export type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | boolean | Promise<void | boolean | Response>

export type buttonProps = {
    label?: string
    tooltip?: string
    showSuccess?: boolean
    toggleOn?: boolean
    subtle?: boolean
    allowRepeat?: boolean
    centerContent?: boolean
    danger?: boolean
    disabled?: boolean
    doSuper?: boolean
    labelOnHover?: boolean
    fillWidth?: boolean
    allowShrink?: boolean
    float?: boolean
    requireConfirm?: boolean
    Left?: ButtonIcon
    Right?: ButtonIcon

    // Style
    squareSize?: number
    fontSize?: string
    textMin?: number

    onClick?: ButtonActionHandler
    onMouseUp?: ButtonActionHandler
    onMouseOver?: ButtonActionHandler
    onMouseLeave?: ButtonActionHandler
    style?: CSSProperties
    setButtonRef?: (ref: HTMLDivElement) => void
}

export type ButtonContentProps = {
    label: string
    Left: ButtonIcon
    Right: ButtonIcon
    setTextWidth: (w: number) => void
    buttonWidth: number
    iconSize: number
    centerContent: boolean
    hidden: boolean
    labelOnHover: boolean
}