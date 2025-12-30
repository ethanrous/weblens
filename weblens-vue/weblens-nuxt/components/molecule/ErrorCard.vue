<template>
    <div
        v-if="error"
        :class="{ 'p-4': true }"
    >
        <div v-if="error.status === 403">
            <h1 class="text-2xl font-bold">Access Forbidden</h1>
            <p class="text-gray-600">You do not have permission to access this resource.</p>
        </div>
        <div v-else>
            <h1 class="text-2xl font-bold">An Unknown Error Occurred</h1>
            <p class="text-gray-600">{{ error }}</p>
            <p class="text-gray-600">Please try again later or file a report</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { WLError } from '~/types/wlError'

defineProps<{
    error?: WLError
}>()
</script>
