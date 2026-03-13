<template>
    <div
        v-if="error"
        :class="{ 'absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 p-4 text-center': true }"
    >
        <h2 class="inline-flex items-center text-2xl font-bold">
            <IconAlertCircle
                v-if="error.status && error.status >= 500"
                :size="28"
                class="mr-1 inline-block text-red-500"
            />
            {{ errorText.header }}
        </h2>
        <p class="text-gray-600">{{ errorText.message }}</p>
    </div>
</template>

<script setup lang="ts">
import { IconAlertCircle } from '@tabler/icons-vue'
import type { WLError } from '~/types/wlError'

const props = defineProps<{
    message?: string
    error?: WLError
}>()

const errorText = computed(() => {
    if (!props.error) return { header: '', message: '' }

    if (props.message) {
        return {
            header: props.message,
            message: props.error.message,
        }
    }

    switch (props.error.status) {
        case 403:
            return {
                header: 'Access Forbidden',
                message: props.error.message ?? 'You do not have permission to access this resource.',
            }
        case 404:
            return {
                header: 'Not Found',
                message: props.error.message ?? 'The requested resource could not be found.',
            }
    }
    return { header: 'An unknown error occurred.', message: props.error.message ?? 'Unknown error' }
})
</script>
