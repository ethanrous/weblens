<template>
    <span
        :class="{
            'inline-flex items-center gap-1 rounded-full py-0.5 text-xs transition': true,
            'px-1.5': compact,
            'px-2': !compact,
            'cursor-pointer': clickable,
            'ring-1 ring-current hover:opacity-75': active,
            'opacity-50 hover:opacity-75': dimmed,
        }"
        :style="{ backgroundColor: tag.color + '30', color: tag.color }"
        @click="$emit('click')"
    >
        {{ tag.name }}
        <IconX
            v-if="removable"
            size="12"
            class="opacity-0 group-hover:opacity-100"
        />
    </span>
</template>

<script setup lang="ts">
import { IconX } from '@tabler/icons-vue'
import type { TagInfo } from '~/stores/tags'

defineProps<{
    tag: TagInfo
    active?: boolean
    dimmed?: boolean
    removable?: boolean
    compact?: boolean
    clickable?: boolean
}>()

defineEmits<{
    (e: 'click'): void
}>()
</script>
