<template>
    <WeblensInput
        :value="value"
        search
        auto-focus
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
}>()
</script>
