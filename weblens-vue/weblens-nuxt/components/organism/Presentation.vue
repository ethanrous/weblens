<template>
    <Teleport to="body">
        <div
            ref="presentation"
            :class="{
                'presentation fullscreen-modal flex-col justify-end sm:flex-row sm:justify-around': true,
            }"
            @click.stop="presentationStore.clearPresentation"
        >
            <div :class="{ 'relative flex h-full w-full': true }">
                <slot
                    name="media"
                    :presentation-size="presentationSize"
                />

                <div
                    :class="{
                        'absolute flex h-full max-w-full shrink-0 flex-col items-center justify-center gap-12 overflow-hidden transition-[width,height,margin] duration-300 lg:relative lg:mb-0 lg:ml-4': true,
                        'w-full p-4 backdrop-blur-xs lg:w-1/3 lg:p-0 lg:backdrop-blur-none': infoOpen,
                        'pointer-events-none opacity-0 lg:w-0 lg:opacity-100': !infoOpen,
                    }"
                >
                    <slot name="fileInfo" />

                    <slot name="mediaInfo" />
                </div>

                <IconInfoCircle
                    :class="{
                        'absolute top-4 right-4 shrink-0 cursor-pointer rounded p-0.5 transition': true,
                        'bg-card-background-primary/50 text-text-primary': infoOpen,
                        'text-text-secondary': !infoOpen,
                    }"
                    size="20"
                    @click.stop="infoOpen = !infoOpen"
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
const infoOpen = ref<boolean>(false)

const presentation = ref<HTMLDivElement>()
const presentationSize = useElementSize(presentation)

const props = defineProps<{
    next: () => void
    previous: () => void
}>()

onKeyUp(['Escape'], (e) => {
    e.stopPropagation()
    presentationStore.clearPresentation()
})

onKeyUp(['i'], (e) => {
    e.stopPropagation()
    infoOpen.value = !infoOpen.value
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
