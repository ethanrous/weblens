<template>
    <div
        ref="stepperContainer"
        :class="{ 'flex h-full w-max shrink-0 items-end gap-0.5': true }"
        @mouseleave="hoverStep = null"
    >
        <div
            v-for="(_, index) in Array.from<number>({ length: stepCount })"
            :key="index"
            :class="{ 'group flex h-full w-2 cursor-pointer items-end': true }"
            @click="emit('select-step', index)"
            @mouseover="handleMouseOver(index)"
        >
            <div
                :class="{
                    'w-2 cursor-pointer group-hover:scale-[1.1]': true,
                    'bg-text-secondary': index <= (hoverStep ?? activeStep),
                    'bg-background-hover': index > (hoverStep ?? activeStep),
                }"
                :style="{
                    height: `${stepSizes[index]}px`,
                }"
            />
        </div>
    </div>
</template>

<script lang="ts" setup>
import { useElementSize } from '@vueuse/core'

const props = defineProps<{
    activeStep: number
    stepCount: number
}>()

const emit = defineEmits<{
    (e: 'select-step', step: number): void
}>()

const hoverStep = ref<number | null>(null)

const stepperContainer = ref<HTMLDivElement>()
const stepperContainerSize = useElementSize(stepperContainer)

const stepSizes = computed(() => {
    const stepSize = stepperContainerSize.height.value / props.stepCount

    return Array.from<number>({ length: props.stepCount }).map((_, index) => {
        return stepSize * (index + 1)
    })
})

function handleMouseOver(index: number) {
    hoverStep.value = index
    //
    // emit('select-step', index)
}
</script>
