<template>
    <div :class="{ 'flex h-full w-full flex-col gap-3 p-4': true }">
        <WeblensCheckbox
            label="Search Recursively"
            :checked="filesStore.searchRecursively"
            @checked:changed="filesStore.setSearchRecurively"
        />
        <span :class="{ 'text-text-tertiary': true }">Tip: Use Shift+{{ keyHintText }} enable recursive search</span>

        <WeblensCheckbox
            label="Search using Regular Expressions"
            :checked="filesStore.searchWithRegex"
            @checked:changed="filesStore.setSearchWithRegex"
        />

        <div
            v-if="tagsStore.tagsList.length > 0"
            class="flex flex-col gap-1.5"
        >
            <span class="text-text-secondary text-xs font-semibold uppercase">Filter by Tags</span>
            <WeblensCheckbox
                v-for="tag in tagsStore.tagsList"
                :key="tag.id"
                :label="tag.name"
                :checked="filesStore.filterTagIDs.has(tag.id)"
                @checked:changed="(checked: boolean) => toggleTagFilter(tag.id, checked)"
            />
        </div>

        <WeblensButton
            :class="{ 'mt-auto ml-auto w-1/3': true }"
            label="Done"
            center-content
            @click="$emit('done')"
        />
    </div>
</template>

<script setup lang="ts">
import useFilesStore from '~/stores/files'
import useTagsStore from '~/stores/tags'
import WeblensCheckbox from '../atom/WeblensCheckbox.vue'
import WeblensButton from '../atom/WeblensButton.vue'
import useLocationStore from '~/stores/location'

const locationStore = useLocationStore()
const filesStore = useFilesStore()
const tagsStore = useTagsStore()

defineEmits<{
    (e: 'done'): void
}>()

const keyHintText = computed(() => {
    if (locationStore.operatingSystem === 'macos') {
        return '⌘K'
    }

    return 'Ctrl+K'
})

function toggleTagFilter(tagID: string, checked: boolean) {
    const newSet = new Set(filesStore.filterTagIDs)
    if (checked) {
        newSet.add(tagID)
    } else {
        newSet.delete(tagID)
    }
    filesStore.setFilterTagIDs(newSet)
}
</script>
