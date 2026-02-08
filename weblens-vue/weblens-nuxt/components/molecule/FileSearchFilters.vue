<template>
    <div :class="{ 'flex h-full w-full flex-col p-4': true }">
        <WeblensCheckbox
            label="Search Recursively"
            :checked="filesStore.searchRecurively"
            @checked:changed="filesStore.setSearchRecurively"
        />
        <span :class="{ 'text-text-tertiary': true }">Tip: Use Shift+{{ keyHintText }} enable recursive search</span>

        <WeblensCheckbox
            label="Search using Regular Expressions"
            :checked="filesStore.searchWithRegex"
            @checked:changed="filesStore.setSearchWithRegex"
        />

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
import WeblensCheckbox from '../atom/WeblensCheckbox.vue'
import WeblensButton from '../atom/WeblensButton.vue'
import useLocationStore from '~/stores/location'

const locationStore = useLocationStore()

const filesStore = useFilesStore()
defineEmits<{
    (e: 'done'): void
}>()

const keyHintText = computed(() => {
    if (locationStore.operatingSystem === 'macos') {
        return '⌘K'
    }

    return 'Ctrl+K'
})
</script>
