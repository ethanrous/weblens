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
            <div class="flex min-h-5 items-center gap-1">
                <span :class="{ 'truncate font-semibold text-nowrap': true }">{{ filename }}</span>
                <div
                    v-if="fileTags.length > 0"
                    class="ml-auto flex shrink-0 gap-0.5"
                    :title="fileTags.map((t) => t.name).join(', ')"
                >
                    <span
                        v-for="tag in fileTags.slice(0, 3)"
                        :key="tag.id"
                        class="h-2 w-2 rounded-full"
                        :style="{ backgroundColor: tag.color }"
                    />
                    <span
                        v-if="fileTags.length > 3"
                        class="text-text-tertiary text-[8px] leading-none"
                    >
                        +{{ fileTags.length - 3 }}
                    </span>
                </div>

                <IconUser
                    v-if="file.shareID"
                    :size="18"
                />
            </div>
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
import { IconUser } from '@tabler/icons-vue'
import useTagsStore from '~/stores/tags'

const tagsStore = useTagsStore()

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

const fileTags = computed(() => {
    return tagsStore.getTagsByFileID(props.file.ID())
})

const fileStats = computed(() => {
    return props.file.FormatSize() + ' - ' + props.file.FormatModified()
})
</script>
