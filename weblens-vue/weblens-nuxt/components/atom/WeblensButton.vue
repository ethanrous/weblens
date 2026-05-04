<template>
    <button
        ref="buttonRef"
        :class="{
            'justify-center': centerContent || !textContent || justIcon,
            'aspect-square': !textContent || justIcon,
            'rounded-none first:rounded-l last:rounded-r': merge === 'row',
            'rounded-none first:rounded-t last:rounded-b': merge === 'column',
            'p-0!': !textContent,
        }"
        :style="{
            height: squareSize ? squareSize + 'px' : undefined,
        }"
        :data-flavor="buttonError ? 'danger' : flavor"
        :data-type="type"
        :data-selected="selected ?? false"
        :data-fill-width="fillWidth"
        :disabled="disabled || doingClick || disabled"
        @click="handleClick"
    >
        <slot />
        <span
            v-if="textContent && !justIcon"
            :class="{
                'mx-1 text-nowrap transition-[width]': true,
            }"
            :style="{
                width: textWidth,
            }"
        >
            {{ textContent }}
        </span>

        <!-- Fake text box just to measure the width -->
        <span
            v-if="textContent"
            ref="fakeText"
            :class="{ 'gone absolute mx-1 text-nowrap': true }"
        >
            {{ textContent }}
        </span>

        <slot name="rightIcon" />
    </button>
</template>

<script setup lang="ts">
import { useElementSize } from '@vueuse/core'
import type { ButtonProps } from '~/types/button'

const { selected = true, allowCollapse, label, onClick, errorText } = defineProps<ButtonProps>()

const slots = useSlots()

const doingClick = ref<boolean>(false)
const buttonError = ref<string>('')

const buttonRef = ref<HTMLButtonElement>()
const buttonSize = useElementSize(buttonRef)

const fakeText = ref<HTMLSpanElement>()
const fakeTextSize = useElementSize(fakeText)

defineExpose({
    click: handleClick,
})

const textWidth = computed(() => {
    if (fakeText.value) {
        return fakeTextSize.width.value + 'px'
    }
    return '0px'
})

const justIcon = computed(() => {
    if (
        allowCollapse &&
        slots.default &&
        buttonSize.width.value < fakeTextSize.width.value + 24 /* Icon size without padding */
    ) {
        return true
    }

    return false
})

const textContent = computed(() => {
    if (buttonError.value) {
        return buttonError.value
    }

    if (label) {
        return label
    }

    return ''
})

async function handleClick(e: MouseEvent) {
    if (!onClick) {
        return
    }

    const maybePromise = onClick(e)
    try {
        if (maybePromise instanceof Promise) {
            doingClick.value = true
            await maybePromise
        }
    } catch (error) {
        console.error('Error during button click:', error)

        if (!errorText) {
            buttonError.value = 'Error'
        } else if (typeof errorText === 'string') {
            buttonError.value = errorText
        } else if (typeof errorText === 'function') {
            buttonError.value = errorText(error as Error)
        }

        await new Promise((resolve) => {
            setTimeout(() => {
                buttonError.value = ''
                resolve(true)
            }, 3000) // Reset error state after 3 seconds
        })
    } finally {
        // Ensure the button is not disabled after the click
        doingClick.value = false
    }
}
</script>
