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
            <div class="flex items-center justify-between">
                <span class="text-text-secondary text-xs font-semibold uppercase">Filter by Tags</span>
                <div class="flex">
                    <WeblensButton
                        label="All"
                        :square-size="24"
                        type="light"
                        :selected="filesStore.filterTagMode === 'and'"
                        merge="row"
                        @click="filesStore.setFilterTagMode('and')"
                    />
                    <WeblensButton
                        label="Any"
                        :square-size="24"
                        type="light"
                        :selected="filesStore.filterTagMode === 'or'"
                        merge="row"
                        @click="filesStore.setFilterTagMode('or')"
                    />
                </div>
            </div>
            <div class="flex flex-wrap gap-1.5">
                <TagPill
                    v-for="tag in tagsStore.tagsList"
                    :key="tag.id"
                    :tag="tag"
                    :active="filesStore.filterTagIDs.has(tag.id!)"
                    :dimmed="!filesStore.filterTagIDs.has(tag.id!) && filesStore.filterTagIDs.size > 0"
                    clickable
                    @click="toggleTagFilter(tag.id!)"
                />
            </div>
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
import WeblensButton from '../atom/WeblensButton.vue'
import useLocationStore from '~/stores/location'
import WeblensCheckbox from '../atom/WeblensCheckbox.vue'
import TagPill from '../atom/TagPill.vue'

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

function toggleTagFilter(tagID: string) {
    const newSet = new Set(filesStore.filterTagIDs)
    if (newSet.has(tagID)) {
        newSet.delete(tagID)
    } else {
        newSet.add(tagID)
    }
    filesStore.setFilterTagIDs(newSet)
}
</script>
