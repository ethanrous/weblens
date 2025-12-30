<template>
    <div
        v-if="!type || type === 'default'"
        :class="{ 'gradient-block funky-spinner': true }"
        :style="style"
    />
    <div
        v-else-if="type === '4square'"
        :class="{ 'grid grid-cols-2 gap-1': true }"
    >
        <div class="gradient-block" />
        <div class="gradient-block" />
        <div class="gradient-block" />
        <div class="gradient-block" />
    </div>
</template>

<script setup lang="ts">
import { toCssUnit } from '~/util/domHelpers'

const props = defineProps<{
    type?: 'default' | '4square'
    size?: string
}>()

const style = computed(() => {
    return {
        width: props.size ? toCssUnit(props.size) : '1rem',
        height: props.size ? toCssUnit(props.size) : '1rem',
    }
})
</script>

<style scoped>
@keyframes spinfunky {
    0% {
        transform: rotate(0deg);
        -webkit-transform: rotate(0deg);
    }
    20% {
        transform: rotate(30deg);
        -webkit-transform: rotate(30deg);
    }
    50% {
        transform: rotate(-380deg);
        -webkit-transform: rotate(-380deg);
    }
    70% {
        transform: rotate(-360deg);
        -webkit-transform: rotate(-360deg);
    }
    100% {
        transform: rotate(-360deg);
        -webkit-transform: rotate(-360deg);
    }
}

.gradient-block {
    height: 1rem;
    width: 1rem;
    border-radius: 0.1em;
    flex-shrink: 0;
    background: linear-gradient(130deg, var(--color-theme-primary), var(--color-bluenova-500));
}

.funky-spinner {
    animation: 3s spinfunky infinite cubic-bezier(0.5, 0, 0.5, 1);
}
</style>
