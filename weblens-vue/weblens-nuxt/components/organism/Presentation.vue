<template>
    <Teleport to="body">
        <div
            ref="presentation"
            :class="{
                'presentation fullscreen-modal flex-col justify-end p-0 sm:flex-row sm:justify-around': true,
            }"
            @click.stop="presentationStore.clearPresentation"
        >
            <div :class="{ 'relative flex h-full w-full': true }">
                <div :class="{ 'mr-auto flex w-full': true }">
                    <slot
                        name="media"
                        :presentation-size="presentationSize"
                    />
                </div>

                <div
                    id="presentation-info-sidecar"
                    :class="{
                        'items-left bg-background-primary/75 absolute flex h-full max-w-full shrink-0 flex-col gap-6 overflow-x-hidden overflow-y-auto p-4 backdrop-blur-xs transition-[width,height,margin,border,padding] duration-300 lg:relative lg:mb-0 lg:ml-3 lg:max-w-125': true,
                        'w-full border lg:w-5/12': presentationStore.infoOpen,
                        'pointer-events-none px-0 opacity-0 lg:w-0 lg:opacity-100': !presentationStore.infoOpen,
                    }"
                    @click.stop.prevent
                >
                    <div
                        v-if="renderInfo"
                        :class="{ 'flex w-max min-w-max flex-col gap-6 px-8 lg:px-0': true }"
                    >
                        <slot name="fileInfo" />

                        <slot name="mediaInfo" />
                    </div>
                </div>

                <IconInfoCircle
                    :class="{
                        'absolute top-4 right-4 shrink-0 cursor-pointer rounded p-0.5 transition': true,
                        'bg-card-background-primary/50 text-text-primary': presentationStore.infoOpen,
                        'text-text-secondary': !presentationStore.infoOpen,
                    }"
                    size="20"
                    @click.stop="presentationStore.infoOpen = !presentationStore.infoOpen"
                />
                <IconArrowLeft
                    :class="{
                        'bg-card-background-primary/50 absolute bottom-10 left-10 m-2 rounded p-1 md:hidden': true,
                    }"
                    size="32"
                    @click.stop="previous()"
                />
                <IconArrowRight
                    :class="{
                        'bg-card-background-primary/50 absolute right-10 bottom-10 m-2 rounded p-1 md:hidden': true,
                    }"
                    size="32"
                    @click.stop="next()"
                />
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { onKeyStroke, onKeyUp, useElementSize } from '@vueuse/core'
import { IconArrowLeft, IconArrowRight, IconInfoCircle } from '@tabler/icons-vue'

const presentationStore = usePresentationStore()

const presentation = ref<HTMLDivElement>()
const presentationSize = useElementSize(presentation)

const props = defineProps<{
    next: () => void
    previous: () => void
}>()

const renderInfo = ref<boolean>(presentationStore.infoOpen)
watchEffect(() => {
    if (presentationStore.infoOpen) {
        renderInfo.value = true
    } else {
        setTimeout(() => {
            if (!presentationStore.infoOpen) renderInfo.value = false
        }, 300)
    }
})

onKeyUp(['Escape'], (e) => {
    e.stopPropagation()
    presentationStore.clearPresentation()
})

onKeyUp(['i'], (e) => {
    e.stopPropagation()
    presentationStore.infoOpen = !presentationStore.infoOpen
})

onKeyStroke(['ArrowRight'], (e) => {
    e.preventDefault()
    e.stopPropagation()
    props.next()
})

onKeyStroke(['ArrowLeft'], (e) => {
    e.preventDefault()
    e.stopPropagation()
    props.previous()
})
</script>
