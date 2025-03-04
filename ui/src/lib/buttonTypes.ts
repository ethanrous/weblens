import { QueryObserverResult } from '@tanstack/react-query'
import { AxiosResponse } from 'axios'
import { CSSProperties, FC } from 'react'

export type ButtonIcon = FC<{ className?: string; size?: number }>

export type ButtonActionPromiseReturn = Promise<
    void | boolean | AxiosResponse | QueryObserverResult
>

export type ButtonActionHandler = (
    e: React.MouseEvent<HTMLElement, MouseEvent>
) => void | ButtonActionPromiseReturn

export type ButtonFlavor = 'default' | 'outline' | 'light'
export type ButtonSize = 'default' | 'tiny' | 'jumbo' | 'small' | 'large'

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

    type?: 'button' | 'submit' | 'reset'
    flavor?: ButtonFlavor
    size?: ButtonSize

    // Style
    squareSize?: number
    fontSize?: string
    textMin?: number

    onClick?: ButtonActionHandler
    onMouseUp?: ButtonActionHandler
    onMouseOver?: ButtonActionHandler
    onMouseLeave?: ButtonActionHandler
    style?: CSSProperties
    className?: string
    setButtonRef?: (ref: HTMLButtonElement) => void
}

export type ButtonContentProps = {
    label: string
    Left: ButtonIcon
    Right: ButtonIcon
    staticTextWidth: number
    setTextWidth: (w: number) => void
    buttonWidth: number
    iconSize: number
    centerContent: boolean
    hidden: boolean
    labelOnHover: boolean
}
