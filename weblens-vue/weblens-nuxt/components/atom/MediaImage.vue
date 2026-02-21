<template>
    <div
        ref="imageContainer"
        :class="{ 'relative flex h-full w-full items-center justify-center': true }"
    >
        <IconExclamationCircle
            v-if="imgError"
            color="red"
        />
        <img
            v-if="media && quality === PhotoQuality.HighRes"
            :class="{
                'media-image-highres absolute bg-center bg-no-repeat': true,
                'object-contain': contain,
                'object-cover': !contain,
                hidden: !highResLoaded,
            }"
            :src="media.ImgUrl(quality)"
            :style="{
                width: imageSize.width,
                height: imageSize.height,
            }"
            @load="highResLoaded = true"
            @click="
                (e) => {
                    if (noClick) {
                        e.stopPropagation()
                    }
                }
            "
        />
        <div
            v-if="media && !highResLoaded && shouldLoad"
            :class="{ 'media-image-lowres animate-fade-in': true }"
            :style="{
                width: imageSize.width,
                height: imageSize.height,
                backgroundImage: `url(${media.ImgUrl()})`,
                backgroundRepeat: 'no-repeat',
                backgroundSize: contain ? 'contain' : 'cover',
                backgroundPosition: 'center',
            }"
            @click="
                (e) => {
                    if (!noClick) {
                        e.stopPropagation()
                        presentationStore.setPresentationMediaID(media.contentID)
                    }
                }
            "
        />

        <Loader
            v-if="quality === PhotoQuality.HighRes && !highResLoaded && shouldLoad"
            :class="{ 'absolute bottom-0 left-0': true }"
        />
    </div>
</template>

<script setup lang="ts">
import { IconExclamationCircle } from '@tabler/icons-vue'
import type WeblensMedia from '~/types/weblensMedia'
import { PhotoQuality } from '~/types/weblensMedia'
import Loader from './Loader.vue'
import { useElementSize, useElementVisibility } from '@vueuse/core'

const presentationStore = usePresentationStore()

const imgError = ref<boolean>(false)
const imageContainer = ref<HTMLDivElement>()
const imageContainerSize = useElementSize(imageContainer)
const highResLoaded = ref<boolean>(false)

const shouldLoad = ref<boolean>(false)
const isVisible = useElementVisibility(imageContainer, { rootMargin: '250px' })
watchEffect(() => {
    if (isVisible.value) {
        shouldLoad.value = true
    }
})

const {
    media = undefined,
    quality = PhotoQuality.LowRes,
    contain = false,
} = defineProps<{
    media?: WeblensMedia
    quality?: PhotoQuality
    contain?: boolean
    placeholder?: boolean
    noClick?: boolean
}>()

const imageSize = computed(() => {
    if (!contain) {
        return {
            width: '100%',
            height: '100%',
        }
    }

    if (!shouldLoad.value || !media || !imageContainerSize.width || !imageContainerSize.height) {
        return {
            width: '',
            height: '',
        }
    }

    if (media.height / media.width > imageContainerSize.height.value / imageContainerSize.width.value) {
        const scaledWidth = (imageContainerSize.height.value / media.height) * media.width
        return {
            width: scaledWidth + 'px',
            height: imageContainerSize.height.value + 'px',
        }
    } else {
        const scaledHeight = (imageContainerSize.width.value / media.width) * media.height
        return {
            width: imageContainerSize.width.value + 'px',
            height: scaledHeight + 'px',
        }
    }
})

defineEmits<{ (e: 'error'): void }>()

watch(
    () => media,
    () => {
        imgError.value = false
        highResLoaded.value = false
    },
)
</script>
