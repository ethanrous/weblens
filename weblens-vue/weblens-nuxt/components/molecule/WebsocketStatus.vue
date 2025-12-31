<template>
    <div
        :class="{
            'h-2 w-2 rounded-[2px]': true,
            'bg-valid': status === 'OPEN',
            'bg-warn': status === 'CONNECTING',
            'bg-danger': status === 'CLOSED',
        }"
    ></div>
</template>

<script setup lang="ts">
import type { WebSocketStatus } from '@vueuse/core'

defineProps<{
    status?: WebSocketStatus
}>()
</script>
