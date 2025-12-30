<template>
    <div
        ref="fileInputContainer"
        :class="{ 'flex w-full': true }"
    >
        <WeblensButton
            :class="{ 'max-md:!rounded': true }"
            v-bind="$props"
            merge="row"
            fill-width
            @click.stop="clickFileInput()"
        >
            <slot />
        </WeblensButton>
        <WeblensButton
            v-if="containerSize.width.value > 100"
            :class="{ 'ml-0.5 hidden md:flex': true }"
            merge="row"
            :disabled="disabled"
            @click="clickFileInput(true)"
        >
            <IconFolder />
        </WeblensButton>

        <input
            ref="fileInput"
            :class="{ haunted: true }"
            type="file"
            multiple
            @change="handleFileChange"
        />
        <input
            ref="fileInputDir"
            :class="{ haunted: true }"
            type="file"
            directory
            webkitdirectory
            @change="handleFileChange"
        />
    </div>
</template>

<script setup lang="ts">
import type { ButtonProps } from '~/types/button'
import WeblensButton from '../atom/WeblensButton.vue'
import { IconFolder } from '@tabler/icons-vue'
import { useElementSize } from '@vueuse/core/index.mjs'

const fileInputContainer = ref<HTMLDivElement>()
const containerSize = useElementSize(fileInputContainer)

const fileInput = ref<HTMLInputElement>()
const fileInputDir = ref<HTMLInputElement>()

defineProps<ButtonProps>()

const emit = defineEmits<{
    (e: 'files-selected', files: FileList): void
}>()

function clickFileInput(isDirectory = false) {
    if (fileInput.value && !isDirectory) {
        fileInput.value.click()
    } else if (fileInputDir.value && isDirectory) {
        fileInputDir.value.click()
    } else {
        console.warn('No file input found')
    }
}

function handleFileChange(e: Event) {
    const files = (e.target as HTMLInputElement).files
    if (files === null) {
        console.warn('No files selected')
        return
    }

    emit('files-selected', files)
}
</script>
