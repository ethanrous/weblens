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
                <tr v-if="rows.length === 0 && emptyText">
                    <td :colspan="columns.length">
                        <span :class="{ 'text-text-secondary p-8 text-center': true }">{{ emptyText }}</span>
                    </td>
                </tr>
                <tr
                    v-for="(row, index) in rows"
                    :key="index"
                    :class="{
                        'hover:bg-muted/50 data-[state=selected]:bg-muted even:bg-accent/25 cursor-pointer border-b transition-colors': true,
                    }"
                >
                    <td
                        v-for="column in columns"
                        :key="column"
                        :class="{
                            'overflow-hidden p-4 text-center align-middle overflow-ellipsis first:text-left last:text-right [&:has([role=checkbox])]:pr-0 *:[[role=checkbox]]:translate-y-0.5': true,
                            'text-nowrap': (row[column] as TableTypes[TableType.JSON])?.tableType !== TableType.JSON,
                        }"
                    >
                        <slot :name="column + '-' + index" />
                        <span v-if="typeof row[column] === 'string' || typeof row[column] === 'number'">
                            {{ row[column] }}
                        </span>
                        <span v-else-if="row[column]?.tableType === TableType.JSON">
                            {{ row[column].value }}
                        </span>
                        <span v-else-if="row[column]?.tableType === TableType.Text">
                            {{ row[column].text }}
                        </span>
                        <WeblensButton
                            v-else-if="row[column]?.tableType === TableType.Button"
                            :class="{ 'h-8': true }"
                            :label="row[column].label"
                            :flavor="row[column].flavor ?? 'primary'"
                            :disabled="row[column].disabled ?? false"
                            @click="row[column].onclick"
                        >
                            <component :is="row[column].icon" />
                        </WeblensButton>
                        <WeblensCheckbox
                            v-else-if="row[column]?.tableType === TableType.Checkbox"
                            :class="{ 'inline-block': true }"
                            :label="row[column].label"
                            :checked="row[column].checked"
                            @checked:changed="
                                (v) => {
                                    const thing = row[column] as TableColumn<TableType.Checkbox>
                                    thing.onchanged(v)
                                }
                            "
                        />
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script setup lang="ts">
// eslint-disable-next-line @typescript-eslint/consistent-type-imports
import { TableType } from '~/types/table'
import type { TableColumn, TableColumns, TableTypes } from '~/types/table'
import WeblensButton from './WeblensButton.vue'
import WeblensCheckbox from './WeblensCheckbox.vue'
import { camelCaseToWords } from '~/util/string'

defineProps<{
    columns: string[]
    rows: TableColumns
    emptyText?: string
}>()
</script>
