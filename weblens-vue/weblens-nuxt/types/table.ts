import type { Icon } from '@tabler/icons-vue'
import type { ButtonProps } from './button'

export enum TableType {
    Button,
    Checkbox,
    JSON,
    Slot,
    Text,
}

type commonCellProps = {
    disabled?: boolean
}

export type TableTypes = {
    [TableType.Button]: {
        tableType: TableType.Button
        icon?: Icon
    } & ButtonProps &
        commonCellProps
    [TableType.Checkbox]: {
        tableType: TableType.Checkbox
        label?: string
        checked: boolean
        onchanged: (value: boolean) => void
    } & commonCellProps
    [TableType.JSON]: {
        tableType: TableType.JSON
        label?: string
        value: Record<string, string | number | boolean | null>
    } & commonCellProps
    [TableType.Slot]: {
        tableType: TableType.Slot
        key: string
        data?: string | number | boolean | Record<string, string | number | boolean | null>
    } & commonCellProps
    [TableType.Text]: {
        tableType: TableType.Text
        text?: string
    } & commonCellProps
}

export type SectionHeaderRow = {
    sectionHeader: true
    text: string
    className?: string
    icon?: Icon
}

export type TableRow = Record<string, TableColumn> | SectionHeaderRow

export type TableColumns = TableRow[]

export type TableColumn<T extends TableType = TableType> = TableTypes[T]
