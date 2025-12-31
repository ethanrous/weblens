<template>
    <div
        ref="containerRef"
        :class="{
            'bg-background-primary goneable absolute top-18 right-2 z-20 flex flex-col items-center overflow-hidden rounded border shadow-md transition-[height,width,border-radius,opacity] duration-300 ease-out': true,
            gone: uploads.length === 0,
            'w-72': open,
            'w-56': !open,
        }"
        :style="{
            height: height,
        }"
        @click="stayOpen = !stayOpen"
    >
        <div
            :class="{
                'z-10 flex w-full shrink-0 items-center border-b py-1 pl-2 transition-[border] duration-300': true,
                'border-b-transparent': !open,
            }"
        >
            <span>{{ totalProgress.soFar }} / {{ totalProgress.total }}</span>
            <ProgressSquare
                :class="{ 'ml-2 h-4 max-w-full min-w-0 grow': true }"
                :progress="totalProgress.percent"
            />
            <IconX
                size="20"
                :class="{ 'clickable m-1 shrink-0': true }"
                @click.stop="uploadStore.clearUploads"
            />
        </div>

        <div
            :class="{
                'goneable no-scrollbar box-border flex w-full max-w-full min-w-0 flex-col overflow-y-auto transition-[height,width,margin,opacity,padding] duration-[inherit]': true,
                'h-max py-1 opacity-100': open,
                'gone m-0': !open,
            }"
        >
            <SingleUploadBox
                v-for="upload of uploads"
                :key="upload.localUploadID + upload.uploadedSoFar"
                :upload="upload"
                :upload-box-open="open"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { useElementHover } from '@vueuse/core'
import ProgressSquare from '../atom/ProgressSquare.vue'
import SingleUploadBox from '../molecule/SingleUploadBox.vue'
import { IconX } from '@tabler/icons-vue'
import { humanBytes } from '~/util/humanBytes'

const containerRef = ref<HTMLElement>()
const hover = useElementHover(containerRef)
const stayOpen = ref(true)

const open = computed(() => {
    return hover.value || stayOpen.value
})

const uploadStore = useUploadStore()

const totalProgress = computed((prev?: { percent: number; soFar: string; total: string }) => {
    let soFar = 0,
        total = 0

    for (const upload of uploadStore.uploads.values()) {
        soFar += upload.uploadedSoFar ?? 0
        total += upload.totalSize ?? 0
    }

    const newVal = {
        percent: (soFar / total) * 100,
        soFar: humanBytes(soFar).join(''),
        total: humanBytes(total).join(''),
    }

    if (prev && prev.soFar === newVal.soFar && prev.total === newVal.total) {
        return prev
    }

    return newVal
})

const height = computed(() => {
    return open.value ? `${Math.min(uploadStore.uploads.size * 80, window.innerHeight * 0.5) + 56}px` : '2.5rem'
})

const uploads = computed(() => {
    const uploads = Array.from(uploadStore.uploads.values())
    uploads.sort((u1, u2) => {
        if (u1.progressPercent === 100 && u2.progressPercent === 100) {
            return u1.startTime - u2.startTime
        }

        if (u1.progressPercent === 100) {
            return 1
        }

        if (u2.progressPercent === 100) {
            return -1
        }

        return u1.startTime - u2.startTime
    })

    return uploads
})
</script>
