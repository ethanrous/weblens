<template>
    <div
        :class="{
            'relative flex shrink-0 overflow-hidden rounded-[2px] transition-[height,width]': true,
        }"
    >
        <div
            :class="{
                'gradient-progress-box': true,
                '!bg-danger': failed,
            }"
            :data-failed="failed"
            :style="{ width }"
        />
        <div
            :class="{
                'gradient-outline-box': true,
                'border-0 !bg-transparent before:opacity-0': showOutline === false,
            }"
        />
    </div>
</template>

<script setup lang="ts">
const {
    progress = 0,
    failed = false,
    showOutline = true,
} = defineProps<{
    progress?: number
    failed?: boolean
    showOutline?: boolean
}>()

const width = computed(() => {
    let prog = progress
    if (!prog || isNaN(prog) || prog < 0) {
        prog = 0
    } else if (prog > 100) {
        prog = 100
    }

    return prog + '%'
})
</script>

<style scoped>
.gradient-outline-box {
    display: flex;
    align-items: center;
    height: 100%;
    width: 100%;

    position: absolute;
    box-sizing: border-box;

    background-color: var(--color-background-primary);
    background-clip: padding-box;
    border: solid 1px transparent;
    border-radius: 0.25em;

    &:before {
        content: '';
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        z-index: -1;
        margin: -1px;
        border-radius: inherit;
        display: flex;
        background: linear-gradient(130deg, var(--color-theme-primary), var(--color-bluenova-500));
    }
}

.gradient-progress-box {
    position: absolute;
    height: 100%;
    background: linear-gradient(130deg, var(--color-theme-primary), var(--color-bluenova-500));
    z-index: 2;

    transition: width 150ms var(--ease-wl-default);
}

.gradient-progress-box[data-failed='true'] {
    background: linear-gradient(130deg, var(--color-danger), var(--color-danger));
}
</style>
