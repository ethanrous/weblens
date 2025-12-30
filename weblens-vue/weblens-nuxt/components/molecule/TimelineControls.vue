<template>
    <div :class="{ 'timeline-controls flex h-full w-max items-center gap-2': true }">
        <SizeStepper
            v-if="locationStore.isInTimeline"
            :class="{ 'hidden lg:flex': true }"
            :active-step="(mediaStore.timelineImageSize - TIMELINE_IMAGE_MIN_SIZE) / 50"
            :step-count="(TIMELINE_IMAGE_MAX_SIZE - TIMELINE_IMAGE_MIN_SIZE) / 50 + 1"
            @select-step="(v) => mediaStore.updateImageSize(v * 50 + TIMELINE_IMAGE_MIN_SIZE)"
        />
        <WeblensButton @click="mediaStore.toggleSortDirection">
            <IconSortAscending v-if="mediaStore.timelineSortDirection === 1" />
            <IconSortDescending v-if="mediaStore.timelineSortDirection === -1" />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import SizeStepper from '../atom/SizeStepper.vue'
import WeblensButton from '../atom/WeblensButton.vue'
import { IconSortAscending, IconSortDescending } from '@tabler/icons-vue'
import useLocationStore from '~/stores/location'

const locationStore = useLocationStore()
const mediaStore = useMediaStore()
</script>
