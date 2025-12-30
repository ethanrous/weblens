<template>
    <div
        class="mx-4 mb-2 flex h-18 max-w-full shrink-0 flex-col items-center gap-1 border-b pb-2 text-nowrap select-none first:mt-2 last:border-none"
    >
        <div :class="{ 'flex w-full items-center gap-1': true }">
            <IconFolder v-if="upload.type === 'folder'" />
            <IconFile v-else-if="upload.type === 'file'" />

            <span :class="{ truncate: true }">{{ upload.name }}</span>
            <IconCheck
                v-if="status === 'completed'"
                :class="{ 'ml-auto': true }"
            />
        </div>
        <div
            v-if="status === 'uploading'"
            :class="{ 'text-text-secondary flex w-full': true }"
        >
            <span :class="{ 'mr-1 border-r pr-1': true }"> {{ upload.progressPercent }}% </span>
            <span :class="{ 'mr-1 border-r pr-1': true }">
                {{ speedStr }}
            </span>
            <span>
                {{ remainingTime }}
            </span>
        </div>
        <span
            v-else-if="status === 'pending'"
            :class="{ 'text-text-secondary w-full': true }"
        >
            Queued... {{ status }}
        </span>
        <span
            v-else-if="status === 'failed'"
            :class="{ 'text-danger w-full': true }"
        >
            Failed
        </span>
        <span
            v-else-if="status === 'completed'"
            :class="{ 'text-text-secondary w-full': true }"
        >
            Finished in {{ humanDuration(upload.endTime - upload.startTime) }}
        </span>

        <ProgressSquare
            :class="{
                'h-2 w-full': true,
            }"
            :failed="status === 'failed'"
            :progress="upload.progressPercent"
        />
    </div>
</template>

<script setup lang="ts">
import type { UploadInfo } from '~/types/uploadTypes'
import { humanBytes, humanDuration } from '~/util/humanBytes'
import ProgressSquare from '../atom/ProgressSquare.vue'
import { IconCheck, IconFile, IconFolder } from '@tabler/icons-vue'

const props = defineProps<{
    upload: UploadInfo
    uploadBoxOpen: boolean
}>()

const speedBytesPerSec = computed(() => {
    if (!props.uploadBoxOpen || props.upload.status !== 'uploading') {
        return 0
    }

    const oldestSample = props.upload.rate[props.upload.rate.length - 1]
    const mostRecentSample = props.upload.rate[0]

    if (!oldestSample || !mostRecentSample) {
        return 0
    }

    const windowMs = mostRecentSample.time - oldestSample.time
    const windowBytes = mostRecentSample.totalBytes - oldestSample.totalBytes

    return (windowBytes / windowMs) * 1000
})

const speedStr = computed(() => {
    if (!props.uploadBoxOpen || props.upload.status !== 'uploading') {
        return '-'
    }

    return humanBytes(speedBytesPerSec.value).join('') + '/s'
})

const remainingTime = computed(() => {
    if (!props.uploadBoxOpen || props.upload.status !== 'uploading') {
        return '-'
    }

    return humanDuration(1000 * ((props.upload.totalSize - props.upload.uploadedSoFar) / speedBytesPerSec.value))
})

const status = computed(() => {
    return props.upload.status
})
</script>
