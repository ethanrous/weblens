<template>
    <div :class="{ 'flex overflow-y-auto rounded border': true }">
        <table
            :class="{
                'w-full caption-bottom border-separate border-spacing-0 text-sm xl:table-fixed': true,
            }"
        >
            <thead :class="{ 'bg-background-primary sticky top-0 z-50 [&_tr]:border-b': true }">
                <tr
                    :class="{
                        'hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors': true,
                    }"
                >
                    <th
                        v-for="column in columns"
                        :key="column"
                        :class="{
                            'text-text-secondary bg-background relative h-10 w-max border-r border-b px-2 align-middle font-medium whitespace-nowrap first:text-left last:border-r-0 last:text-right [&:has([role=checkbox])]:pr-0 *:[[role=checkbox]]:translate-y-0.5': true,
                        }"
                    >
                        {{ camelCaseToWords(column) }}
                    </th>
                </tr>
            </thead>
            <tbody :class="{ '[&_tr:last-child]:border-0': true }">
                <template
                    v-for="(item, index) in rows"
                    :key="index"
                >
                    <tr
                        v-if="item && 'sectionHeader' in item"
                        :class="[item.className || '']"
                    >
                        <td
                            :colspan="columns.length"
                            :class="{ 'p-4 font-medium': true }"
                        >
                            <div class="flex items-center justify-center gap-2">
                                <component
                                    v-if="item.icon"
                                    :is="item.icon"
                                />
                                {{ item.text }}
                            </div>
                        </td>
                    </tr>
                    <template v-else-if="!isEmptyRow(item)">
                        <tr
                            :class="{
                                'hover:bg-muted/50 data-[state=selected]:bg-muted even:bg-accent/25 border-b transition-colors': true,
                            }"
                        >
                            <td
                                v-for="column in columns"
                                :key="column"
                                :class="{
                                    'overflow-hidden p-4 text-center align-middle first:text-left last:text-right [&:has([role=checkbox])]:pr-0 *:[[role=checkbox]]:translate-y-0.5': true,
                                    'text-nowrap':
                                        (item[column] as TableColumn<TableType.JSON>)?.tableType !== TableType.JSON,
                                }"
                            >
                                <TableCell
                                    :column="column"
                                    :index="index"
                                    :cell-data="item[column]"
                                />
                            </td>
                        </tr>
                    </template>
                </template>
            </tbody>
        </table>
    </div>
</template>

<script setup lang="ts">
// eslint-disable-next-line @typescript-eslint/consistent-type-imports
import { TableType } from '~/types/table'
import type { TableColumn, TableRow } from '~/types/table'
import TableCell from './TableCell.vue'
import { camelCaseToWords } from '~/util/string'

const props = defineProps<{
    columns: string[]
    rows: TableRow[]
    emptyText?: string
}>()

function isEmptyRow(row: TableRow): boolean {
    if ('sectionHeader' in row) return true
    if (row === null || row === undefined) return true
    return Object.keys(row).length === 0
}
</script>
