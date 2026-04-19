<template>
    <slot :name="column + '-' + index" />
    <span v-if="typeof cellData === 'string' || typeof cellData === 'number'">
        {{ cellData }}
    </span>
    <span v-else-if="cellData?.tableType === TableType.JSON">
        {{ cellData.value }}
    </span>
    <span v-else-if="cellData?.tableType === TableType.Text">
        {{ cellData.text }}
    </span>
    <WeblensButton
        v-else-if="cellData?.tableType === TableType.Button"
        v-bind="cellData"
    >
        <component
            :is="cellData.icon"
            v-if="cellData.icon"
        />
    </WeblensButton>
    <WeblensCheckbox
        v-else-if="cellData?.tableType === TableType.Checkbox"
        :class="{ 'inline-block': true }"
        :label="cellData.label"
        :checked="cellData.checked"
        :disabled="cellData.disabled"
        @checked:changed="
            (v) => {
                const thing = cellData as TableColumn<TableType.Checkbox>
                thing.onchanged(v)
            }
        "
    />
</template>

<script setup lang="ts">
import { TableType } from '~/types/table'
import type { TableColumn } from '~/types/table'
import WeblensButton from './WeblensButton.vue'
import WeblensCheckbox from './WeblensCheckbox.vue'

defineProps<{
    column: string
    index: number
    cellData: TableColumn | string | number | null | undefined
}>()
</script>
