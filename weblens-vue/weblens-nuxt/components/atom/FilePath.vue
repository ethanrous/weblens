<template>
    <span
        v-for="(part, i) of parts"
        :key="i"
        :class="{ 'inline-flex items-center': true }"
    >
        {{ part }}
        <IconChevronRight
            v-if="i < parts.length - 1"
            size="14"
        />
    </span>
</template>

<script setup lang="ts">
import { IconChevronRight } from '@tabler/icons-vue'
import type { PortablePath } from '~/types/portablePath'

const props = defineProps<{
    path: PortablePath
    omitLast?: boolean
}>()

const parts = computed(() => {
    const parts = props.path.parts
    parts.splice(0, 1) // Remove the first part

    if (props.omitLast) {
        parts.pop() // Remove the last part if omitLast is true
    }

    return parts
})
</script>
