<template>
    <div :class="{ 'absolute flex h-96 w-full p-4': true }">
        <WeblensCheckbox
            label="Show Raws"
            :checked="mediaStore.showRaw"
            @changed="mediaStore.setShowRaw"
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
import WeblensCheckbox from '../atom/WeblensCheckbox.vue'
import WeblensButton from '../atom/WeblensButton.vue'

const mediaStore = useMediaStore()
defineEmits<{
    (e: 'done'): void
}>()
</script>
