<template>
    <div
        ref="dividerRef"
        :class="{ 'relative flex items-center py-1 pt-4': true }"
    >
        <div
            :class="{
                'border-border-primary flex-grow border-t': true,
                'flex-1/2': labelJustify === 'right',
            }"
        />
        <span
            v-if="label && dividerSize.width.value > 60"
            :class="{ 'text-border-secondary mx-3 flex-shrink select-none': true }"
        >
            {{ label }}
        </span>
        <div
            :class="{
                'border-border-primary flex-grow border-t': true,
                'flex-1/2': labelJustify === 'left',
            }"
        />
    </div>
</template>

<script setup lang="ts">
import { useElementSize } from '@vueuse/core'

const dividerRef = ref<HTMLDivElement>()
const dividerSize = useElementSize(dividerRef)

defineProps<{
    label?: string
    labelJustify?: 'left' | 'center' | 'right'
}>()
</script>
