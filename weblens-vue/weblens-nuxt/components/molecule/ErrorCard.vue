<template>
    <div
        v-if="error"
        :class="{ 'absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 p-4 text-center': true }"
    >
        <h2 class="text-2xl font-bold">{{ errorText.header }}</h2>
        <p class="text-gray-600">{{ errorText.message }}</p>
    </div>
</template>

<script setup lang="ts">
import type { WLError } from '~/types/wlError'

const props = defineProps<{
    error?: WLError
}>()

const errorText = computed(() => {
    if (!props.error) return { header: '', message: '' }
    switch (props.error.status) {
        case 403:
            return {
                header: 'Access Forbidden',
                message: 'You do not have permission to access this resource.',
            }
        case 404:
            return {
                header: 'Not Found',
                message: 'The requested resource could not be found.',
            }
    }
    return { header: 'An unknown error occurred.', message: props.error.toString() }
})
</script>
