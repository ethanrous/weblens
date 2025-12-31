<template>
    <div
        :class="{ 'group relative flex h-full max-h-max w-full max-w-max items-center justify-center': true }"
        @click.stop
    >
        <video
            ref="videoRef"
            :class="{ 'max-h-full object-contain': true }"
            autoplay
            :poster="media.ImgUrl()"
            @loadeddata="onLoadedData"
            @click.stop="handleClick"
        />

        <div
            :class="{ 'playbar-container group-hover:!opacity-100': true }"
            :style="{ width: videoSize.width + 'px' }"
        >
            <Seeker
                :percent="playbackPercent"
                @seek="handleSeek"
            />
        </div>

        <IconPlayerPlayFilled
            size="60"
            :class="{
                'bg-card-background-primary/35 absolute cursor-pointer rounded p-1 shadow transition duration-[200ms]': true,
                gone: !paused,
            }"
            @click="handleClick"
        />
        <Loader
            v-if="!failed && (waiting || seeking)"
            :class="{ absolute: true }"
        />
        <IconExclamationCircle
            v-if="failed"
            color="red"
            size="64"
            :class="{
                'bg-card-background-primary/25 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded p-2': true,
            }"
        />
    </div>
</template>

<script setup lang="ts">
import { useElementSize, useMagicKeys, useMediaControls } from '@vueuse/core'
import Hls, { Events, type ErrorData } from 'hls.js'
import HlsWorkerUrl from 'hls.js/dist/hls.worker.js?url'
import type WeblensMedia from '~/types/weblensMedia'
import { IconExclamationCircle, IconPlayerPlayFilled } from '@tabler/icons-vue'
import Loader from '../atom/Loader.vue'
import Seeker from '../atom/Seeker.vue'

const loaded = ref<boolean>(false)
const failed = ref<boolean>(false)
const videoRef = ref<HTMLVideoElement | null>(null)
const videoSize = useElementSize(videoRef)

const { playing, waiting, seeking, currentTime } = useMediaControls(videoRef)
const hls = shallowRef<Hls | undefined>(undefined)

const props = defineProps<{
    media: WeblensMedia
}>()

const paused = computed(() => {
    return !playing.value && !waiting.value && !seeking.value
})

const playbackPercent = computed(() => {
    if (!props.media.duration || !currentTime.value) {
        return 0
    }

    return ((currentTime.value * 1000) / props.media.duration) * 100
})

watchEffect(() => {
    if (!videoRef.value) {
        return
    }

    if (videoRef.value!.canPlayType('application/vnd.apple.mpegurl')) {
        console.debug('Native playback is supported. Using native HLS playback')

        videoRef.value!.src = props.media.StreamVideoUrl()
    } else if (Hls.isSupported()) {
        hls.value = new Hls({
            testBandwidth: false,
            workerPath: HlsWorkerUrl,
            enableWorker: true,
            lowLatencyMode: true,
            backBufferLength: 90,
            maxBufferLength: 600,
            maxMaxBufferLength: 600,
            maxBufferSize: 600 * 1000 * 1000, // 600MB
        })
    } else {
        console.error('native playback and HLS.js are both not supported in this browser')

        failed.value = true
    }
})

function onHlsEror(_event: typeof Hls.Events.ERROR, data: ErrorData): void {
    if (data.fatal) {
        failed.value = true
        console.error('HLS FATAL error', _event, data)
    } else {
        console.warn('HLS non-fatal error', _event, data)
    }
    // if (data.fatal && hls) {
    //     switch (data.type) {
    //         case ErrorTypes.NETWORK_ERROR: {
    //             // Try to recover network error
    //             useSnackbar(t('networkError'), 'error')
    //             console.error('fatal network error encountered, try to recover')
    //             hls.startLoad()
    //             break
    //         }
    //         case ErrorTypes.MEDIA_ERROR: {
    //             useSnackbar(t('mediaError'), 'error')
    //             console.error('fatal media error encountered, try to recover')
    //             hls.recoverMediaError()
    //             break
    //         }
    //         default: {
    //             /**
    //              * Can't recover from unknown errors
    //              */
    //             useSnackbar(t('cantPlayItem'), 'error')
    //             playbackManager.stop()
    //             break
    //         }
    //     }
    // }
}

async function onLoadedData(): Promise<void> {
    if (videoRef.value) {
        /**
         * Makes the resume start from the correct time
         */
        loaded.value = true
        if (videoRef.value.paused) {
            videoRef.value.play()
        }
        // videoRef.value.currentTime = 0
    }
}

const videoUrl = computed(() => props.media.StreamVideoUrl())

const { space } = useMagicKeys()

watch(space!, (isPressed) => {
    if (isPressed && videoRef.value) {
        handleClick()
    }
})

function handleClick(): void {
    if (videoRef.value) {
        if (videoRef.value.paused) {
            videoRef.value.play()
        } else {
            videoRef.value.pause()
        }
    }
}

watch(
    videoUrl,
    async (newUrl) => {
        if (hls.value) {
            hls.value.stopLoad()
        }

        /**
         * Ensure element is mounted before setting the source.
         */
        await nextTick()

        if (videoRef.value && (!newUrl || !hls.value)) {
            /**
             * For the video case, Safari iOS doesn't support hls.js but supports native HLS.
             *
             * We stringify undefined instead of skipping this block when there's no new source url,
             * so the player doesn't restart playback of the previous item
             */
            videoRef.value.src = String(newUrl)
        } else if (hls.value && newUrl) {
            /**
             * We need to check if HLS.js can handle transcoded audio to remove the video check
             */
            hls.value.loadSource(newUrl)
        }
    },
    { immediate: true },
)

function detachHls(): void {
    if (hls.value) {
        hls.value.detachMedia()
        hls.value.off(Events.ERROR, onHlsEror)
    }
}

function handleSeek(percent: number): void {
    if (videoRef.value) {
        const duration = props.media.duration || 0
        const seekTime = (percent / 100) * duration
        videoRef.value.currentTime = seekTime / 1000 // Convert to seconds
    }
}

watch(videoRef, () => {
    detachHls()

    if (videoRef.value) {
        if (hls.value) {
            hls.value.attachMedia(videoRef.value)
            hls.value.on(Events.ERROR, onHlsEror)
        }
    }
})

onScopeDispose(() => {
    detachHls()
    hls.value?.destroy()
})
</script>

<style lang="css" scoped>
.playbar-container {
    display: flex;
    align-items: center;
    position: absolute;
    bottom: calc(-1px);
    height: 2rem;
    width: 100%;
    padding: 0 0.5rem;
    background: linear-gradient(0deg, rgba(0, 0, 0, 0.7) 0%, rgba(0, 0, 0, 0) 100%);
    opacity: 0;
    transition: opacity 200ms var(--ease-wl-default);
}
</style>
