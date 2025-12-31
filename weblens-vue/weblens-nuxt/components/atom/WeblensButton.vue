<template>
    <button
        ref="buttonRef"
        :class="{
            'animate-fade-in': true,
            'justify-center': centerContent || !textContent || justIcon,
            'aspect-square': !textContent || justIcon,
            'rounded-none first:rounded-l last:rounded-r': merge === 'row',
            'rounded-none first:rounded-t last:rounded-b': merge === 'column',
            '!p-0': !textContent,
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

const props = defineProps<ButtonProps>()

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
        props.allowCollapse &&
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

    if (props.label) {
        return props.label
    }

    return ''
})

async function handleClick(e: MouseEvent) {
    if (!props.onClick) {
        return
    }

    const maybePromise = props.onClick(e)
    try {
        if (maybePromise instanceof Promise) {
            doingClick.value = true
            await maybePromise
        }
    } catch (error) {
        console.error('Error during button click:', error)

        if (!props.errorText) {
            buttonError.value = 'Error'
        } else if (typeof props.errorText === 'string') {
            buttonError.value = props.errorText
        } else if (typeof props.errorText === 'function') {
            buttonError.value = props.errorText(error as Error)
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
