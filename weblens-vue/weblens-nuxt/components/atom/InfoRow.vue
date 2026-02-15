<template>
    <div class="flex items-center gap-3 py-1.5">
        <div class="text-text-secondary mt-0.5 shrink-0">
            <slot name="icon" />
        </div>
        <div class="flex min-w-0 flex-col">
            <span class="text-text-secondary text-xs">{{ label }}</span>
            <slot>
                <span class="text-text-primary text-sm">{{ value }}</span>
            </slot>
        </div>
    </div>
</template>

<script setup lang="ts">
defineProps<{
    label: string
    value?: string
}>()
</script>
