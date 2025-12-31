<template>
    <div
        :class="{
            'bg-card-background-primary/75 inline-flex h-9 shrink-0 cursor-text items-center gap-1 overflow-hidden rounded border transition': true,
            'select-all': canCopy,
            'select-none': !canCopy,
            'border-green-600': copied,
        }"
        @click.stop="copyToClipboard"
    >
        <div
            :class="{
                'ml-1': true,
                'flex aspect-square h-12 max-h-full items-center justify-center border-r pr-1': slots.default,
            }"
        >
            <slot />
        </div>
        <span :class="{ 'block max-w-full min-w-0 truncate': true }">
            {{ text ?? '...' }}
        </span>
        <IconClipboard
            :class="{
                'ml-auto shrink-0 rounded p-0.5': true,
                'hover:bg-card-background-hover cursor-pointer': canCopy,
                'text-text-tertiary': !canCopy,
            }"
        />
    </div>
</template>

<script setup lang="ts">
import { IconClipboard } from '@tabler/icons-vue'
import { useTimeoutFn } from '@vueuse/core'
const slots = useSlots()
const isSecure = useSecureContext()

const props = defineProps<{
    text?: string
}>()

const canCopy = computed(() => {
    return isSecure && props.text
})

// TODO: This seems dumb.
const { isPending: copied, start: doCopied } = useTimeoutFn(() => {}, 1000, {
    immediate: false,
    immediateCallback: true,
})

async function copyToClipboard() {
    if (!canCopy.value) {
        return
    }

    await navigator.clipboard.writeText(props.text!)

    doCopied()
}
</script>
