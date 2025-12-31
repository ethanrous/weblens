<template>
    <div
        :class="{
            'relative flex h-full w-full cursor-pointer rounded border transition': true,
        }"
    >
        <div :class="{ 'p-1.5': true }">
            <div :class="{ 'aspect-square h-full overflow-hidden rounded': true }">
                <slot name="file-visual" />
            </div>
        </div>
        <div
            :class="{
                'flex h-full w-full items-center gap-0.5 px-2 pb-1 select-none': true,
            }"
        >
            <span :class="{ 'h-max truncate text-lg font-semibold text-nowrap': true }">{{ file.GetFilename() }}</span>
            <span
                :class="{
                    'mt-auto ml-auto text-xs': true,
                    'text-text-secondary': !fileState.Has(SelectedState.Moved),
                }"
            >
                {{ file.FormatSize() + ' - ' + file.FormatModified() }}
            </span>
            <span
                :class="{ 'absolute top-4 right-4 inline-block text-center leading-none sm:hidden': true }"
                @click="(e) => $emit('contextMenu', e)"
            >
                ...
            </span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { SelectedState } from '@/types/weblensFile'
import type WeblensFile from '@/types/weblensFile'

defineProps<{
    file: WeblensFile
    fileState: SelectedState
}>()

defineEmits<{
    (e: 'contextMenu', event: MouseEvent): void
}>()
</script>
