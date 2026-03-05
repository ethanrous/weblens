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
            <div
                v-if="fileTags.length > 0"
                class="flex items-center gap-1"
            >
                <span
                    v-for="tag in fileTags.slice(0, 3)"
                    :key="tag.id"
                    class="inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-xs"
                    :style="{ backgroundColor: tag.color + '30', color: tag.color }"
                >
                    {{ tag.name }}
                </span>
                <span
                    v-if="fileTags.length > 3"
                    class="text-text-tertiary text-xs"
                >
                    +{{ fileTags.length - 3 }}
                </span>
            </div>
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
import useTagsStore from '~/stores/tags'

const tagsStore = useTagsStore()

const props = defineProps<{
    file: WeblensFile
    fileState: SelectedState
}>()

const fileTags = computed(() => {
    return tagsStore.getTagsByFileID(props.file.ID())
})

defineEmits<{
    (e: 'contextMenu', event: MouseEvent): void
}>()
</script>
