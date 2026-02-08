import type { Icon } from '@tabler/icons-vue'
import type { ButtonProps } from './button'

export type TableColumns = Record<string, TableColumn>[]

export enum TableType {
    Button,
    Checkbox,
    JSON,
    Slot,
    Text,
}

export type TableTypes = {
    [TableType.Button]: {
        tableType: TableType.Button
        onclick?: (e: Event) => void
        icon?: Icon
    } & ButtonProps
    [TableType.Checkbox]: {
        tableType: TableType.Checkbox
        label?: string
        checked: boolean
        onchanged: (value: boolean) => void
    }
    [TableType.JSON]: {
        tableType: TableType.JSON
        label?: string
        value: Record<string, string | number | boolean | null>
    }
    [TableType.Slot]: {
        tableType: TableType.Slot
        key: string
        data?: string | number | boolean | Record<string, string | number | boolean | null>
    }
    [TableType.Text]: {
        tableType: TableType.Text
        text?: string
    }
}

export type TableColumn<T extends TableType = TableType> = TableTypes[T]
