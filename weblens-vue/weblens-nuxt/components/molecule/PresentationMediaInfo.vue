<template>
    <template v-if="media">
        <div class="flex w-full flex-col">
            <span class="text-text-secondary mb-1 text-xs font-semibold uppercase"> Media Details </span>

            <InfoRow
                v-if="formattedDate"
                label="Created"
                :value="formattedDate"
            >
                <template #icon>
                    <IconCalendarEvent size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="media.width && media.height"
                label="Dimensions"
                :value="dimensions"
            >
                <template #icon>
                    <IconAspectRatio size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="formatName"
                label="Format"
            >
                <template #icon>
                    <IconPhoto size="18" />
                </template>
                <span class="text-text-primary flex items-center gap-1.5 text-sm">
                    {{ formatName }}
                    <span
                        v-if="mediaType?.IsRaw"
                        class="bg-card-background-secondary rounded px-1.5 py-0.5 text-xs font-semibold"
                    >
                        RAW
                    </span>
                </span>
            </InfoRow>

            <WeblensButton
                :class="{ 'mt-3': true }"
                label="Download JPEG ..."
                type="outline"
                :disabled="qualitySliderOpen"
                @click="
                    () => {
                        quality = 85
                        qualitySliderOpen = !qualitySliderOpen
                    }
                "
            >
                <IconPhotoDown size="20" />
            </WeblensButton>

            <div
                :class="{
                    'overflow-hidden rounded border transition-[border,height,margin]': true,
                    'h-0 border-transparent': !qualitySliderOpen,
                    'mt-3 h-20': qualitySliderOpen,
                }"
            >
                <div :class="{ 'flex h-20 min-h-20 flex-col p-2': true }">
                    <span :class="{ 'text-text-secondary mb-1 inline-flex items-center gap-2': true }">
                        <IconX
                            size="16"
                            :class="{ 'hover:text-text-primary cursor-pointer': true }"
                            @click="qualitySliderOpen = false"
                        />
                        JPEG Quality
                    </span>

                    <div :class="{ 'flex items-center gap-2': true }">
                        <span :class="{ 'min-w-[2.5em]': true }"> {{ quality }}% </span>
                        <Seeker
                            :percent="quality"
                            :do-debounce="false"
                            :always-show-handle="true"
                            @update:percent="(v) => (quality = Math.round(v))"
                        />

                        <WeblensButton
                            :class="{ 'ml-auto': true }"
                            label="Download"
                            type="outline"
                            :square-size="32"
                            @click="handleDownload()"
                        />
                    </div>
                </div>
            </div>

            <InfoRow
                v-if="media.IsVideo() && media.duration"
                label="Duration"
                :value="humanDuration(media.duration)"
            >
                <template #icon>
                    <IconClock size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="media.IsPDF() && media.pageCount > 1"
                label="Pages"
                :value="String(media.pageCount)"
            >
                <template #icon>
                    <IconFiles size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="media.likedBy?.length"
                label="Liked by"
                :value="likedByText"
            >
                <template #icon>
                    <IconHeart size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="media.hidden"
                label="Visibility"
                value="Hidden"
            >
                <template #icon>
                    <IconEyeOff size="18" />
                </template>
            </InfoRow>
        </div>

        <div class="flex w-full flex-col">
            <span class="text-text-secondary mb-1 text-xs font-semibold uppercase"> Share </span>
            <CopyBox
                :class="{
                    'relative w-full min-w-0 overflow-x-auto': true,
                }"
                :text="media.MediaUrl()"
            >
                <IconLink
                    size="20"
                    class="shrink-0"
                />
            </CopyBox>
        </div>

        <div
            v-if="media.location && media.location[0] !== 0"
            class="flex w-full flex-col"
        >
            <span class="text-text-secondary mb-1 text-xs font-semibold uppercase"> Location </span>
            <Mapbox
                :coords="media.location"
                class="h-48 w-full"
            />
        </div>
    </template>
</template>

<script setup lang="ts">
import {
    IconAspectRatio,
    IconCalendarEvent,
    IconClock,
    IconEyeOff,
    IconFiles,
    IconHeart,
    IconLink,
    IconPhoto,
    IconPhotoDown,
    IconX,
} from '@tabler/icons-vue'
import CopyBox from './CopyBox.vue'
import Mapbox from '../atom/Mapbox.vue'
import InfoRow from '../atom/InfoRow.vue'
import { humanDuration } from '~/util/humanBytes'
import WeblensButton from '../atom/WeblensButton.vue'
import { downloadSingleFile } from '~/api/FileBrowserApi'
import useFilesStore from '~/stores/files'
import Seeker from '../atom/Seeker.vue'

const mediaStore = useMediaStore()
const qualitySliderOpen = ref<boolean>(false)
const quality = ref<number>(85)

const props = defineProps<{
    mediaId: string
}>()

const media = computed(() => {
    return mediaStore.mediaMap.get(props.mediaId)
})

const mediaType = computed(() => {
    return media.value?.GetMediaType()
})

const formatName = computed(() => {
    return mediaType.value?.FriendlyName
})

const formattedDate = computed(() => {
    if (!media.value?.createDate) return undefined
    return new Date(media.value.createDate).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
    })
})

const dimensions = computed(() => {
    const m = media.value
    if (!m?.width || !m?.height) return ''
    const mp = ((m.width * m.height) / 1_000_000).toFixed(1)
    return `${m.width} x ${m.height} (${mp} MP)`
})

const likedByText = computed(() => {
    const liked = media.value?.likedBy
    if (!liked?.length) return ''
    if (liked.length === 1) return liked[0]
    return `${liked[0]} and ${liked.length - 1} other${liked.length - 1 > 1 ? 's' : ''}`
})

async function handleDownload() {
    if (!media.value?.fileIDs?.length) return

    const file = useFilesStore().getFileByID(media.value.fileIDs[0])
    let filename = file?.GetFilename()
    if (!filename) {
        filename = `media_${media.value.ID()}`
    }
    if (quality.value !== 100) {
        const parts = filename.split('.')
        filename = `${parts[0]}_${quality.value}.${parts[1] ?? 'jpeg'}`
    }

    await downloadSingleFile(media.value.fileIDs[0], filename, 'jpeg', quality.value)
    qualitySliderOpen.value = false
}
</script>
