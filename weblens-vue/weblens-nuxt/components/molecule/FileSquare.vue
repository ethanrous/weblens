<template>
    <div
        :class="{
            'flex aspect-square h-full w-full cursor-pointer flex-col rounded transition': true,
        }"
    >
        <div
            :class="{
                'flex aspect-square h-full min-h-0 w-full min-w-0 items-center p-1.5 transition-[padding] sm:p-3': true,
            }"
        >
            <div :class="{ 'h-full w-full overflow-hidden rounded': true }">
                <slot name="file-visual" />
            </div>
        </div>
        <div
            :class="{
                'flex h-[15%] min-h-max justify-end gap-0.5 px-2 pb-1 select-none sm:min-h-12 sm:flex-col sm:pb-2': true,
            }"
            :title="filename + ' - ' + fileStats"
        >
            <span :class="{ 'min-h-5 truncate font-semibold text-nowrap': true }">{{ filename }}</span>
            <span
                :class="{
                    'hidden text-xs sm:inline-block': true,
                    'text-text-secondary truncate': !fileState.Has(SelectedState.Moved),
                }"
            >
                {{ fileStats }}
            </span>
            <span
                :class="{ 'ml-auto inline-block text-center leading-none sm:hidden': true }"
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

const props = defineProps<{
    file: WeblensFile
    fileState: SelectedState
}>()

defineEmits<{
    (e: 'contextMenu', event: MouseEvent): void
}>()

const filename = computed(() => {
    return props.file.GetFilename()
})

const fileStats = computed(() => {
    return props.file.FormatSize() + ' - ' + props.file.FormatModified()
})
</script>
