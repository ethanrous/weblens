<template>
    <div
        :class="{
            'inline-flex items-center rounded-xs': true,
            'bg-valid': wsStatus === 'OPEN',
            'bg-warn': wsStatus === 'CONNECTING',
            'bg-danger': wsStatus === 'CLOSED',
        }"
        :style="{ width: size + 'px', height: size + 'px' }"
    >
        <span
            v-if="showAsText"
            :class="{
                'text-valid': wsStatus === 'OPEN',
                'text-warn': wsStatus === 'CONNECTING',
                'text-danger': wsStatus === 'CLOSED',
            }"
            :style="{
                fontSize: size + 4 + 'px',
                marginLeft: size * 1.3 + 'px',
            }"
        >
            {{ wsStatus === 'OPEN' ? 'Connected' : wsStatus === 'CONNECTING' ? 'Connecting...' : 'Offline' }}
        </span>
    </div>
</template>

<script setup lang="ts">
import type { WebSocketStatus } from '@vueuse/core'

const {
    wsStatus = 'CLOSED',
    showAsText = false,
    size = 8,
} = defineProps<{
    wsStatus?: WebSocketStatus
    showAsText?: boolean
    size?: number
}>()
</script>
