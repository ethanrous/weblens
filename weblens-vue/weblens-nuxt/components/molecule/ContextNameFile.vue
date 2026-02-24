<template>
    <WeblensInput
        :value="value"
        search
        auto-focus
        :disabled="disabled"
        @submit="handleSubmit"
    />
</template>

<script setup lang="ts">
import WeblensInput from '../atom/WeblensInput.vue'

const emit = defineEmits<{
    (e: 'submit', value: string): void
}>()

function handleSubmit(value: string) {
    emit('submit', value)
}

defineProps<{
    value?: string
    disabled?: boolean
}>()
</script>
