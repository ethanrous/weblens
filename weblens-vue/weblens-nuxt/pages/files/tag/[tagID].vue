<template>
    <div
        id="filebrowser-container"
        :class="{ 'filebrowser-container relative flex h-full min-h-0 w-full min-w-60 items-center': true }"
    >
        <div
            v-if="loading"
            :class="{ 'm-auto': true }"
        >
            <Loader :class="{ 'h-8 w-8': true }" />
        </div>

        <span
            v-else-if="error"
            :class="{ 'border-warn bg-warn/50 mx-auto max-w-1/2 rounded-sm border p-2 text-center text-xl': true }"
        >
            Failed to load tag files
            <span :class="{ 'text-text-secondary mt-1 text-xs': true }">
                {{ (error as Error)?.message ?? 'Unknown error' }}
            </span>
        </span>

        <div
            v-else-if="tagFiles.length > 0"
            :class="{ 'flex h-full w-full': true }"
        >
            <FileScroller
                :files="tagFiles"
                :no-require-parent-match="true"
            />
        </div>

        <span
            v-else
            :class="{ 'text-text-tertiary m-auto text-center': true }"
        >
            No files tagged with "{{ activeTag?.name }}"
        </span>
    </div>
</template>

<script setup lang="ts">
import Loader from '~/components/atom/Loader.vue'
import FileScroller from '~/components/organism/FileScroller.vue'
import { useWeblensAPI } from '~/api/AllApi'
import useTagsStore from '~/stores/tags'
import WeblensFile from '~/types/weblensFile'

const route = useRoute()
const tagsStore = useTagsStore()

const tagID = computed(() => route.params.tagID as string)

const activeTag = computed(() => tagsStore.tags.get(tagID.value))

const {
    data: tagFilesRaw,
    error,
    status,
} = useAsyncData(
    () => 'tag-files-' + tagID.value,
    async () => {
        const { data } = await useWeblensAPI().TagsAPI.getFilesByTag(tagID.value)
        return data
    },
    { watch: [tagID] },
)

const loading = computed(() => status.value === 'pending')

const tagFiles = computed(() => {
    if (!tagFilesRaw.value) return []
    return tagFilesRaw.value.map((f) => new WeblensFile(f))
})

onMounted(() => {
    if (tagsStore.tagsList.length === 0) {
        tagsStore.fetchTags()
    }
})
</script>
