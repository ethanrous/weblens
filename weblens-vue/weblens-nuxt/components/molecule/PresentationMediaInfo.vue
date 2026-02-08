<template>
    <template v-if="media">
        <CopyBox
            :class="{
                'relative mb-2 w-full min-w-0 overflow-x-auto': true,
            }"
            :text="media.MediaUrl()"
        >
            <IconPhoto
                size="20"
                :class="{ 'shrink-0': true }"
            />
        </CopyBox>

        <Mapbox
            v-if="media.location && media.location[0] !== 0"
            :coords="media.location"
            :class="{ 'h-98 w-full': true }"
        />
    </template>
</template>

<script setup lang="ts">
import CopyBox from './CopyBox.vue'
import { IconPhoto } from '@tabler/icons-vue'
import Mapbox from '../atom/Mapbox.vue'

const mediaStore = useMediaStore()

const props = defineProps<{
    mediaId: string
}>()

const media = computed(() => {
    return mediaStore.mediaMap.get(props.mediaId)
})
</script>
