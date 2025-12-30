<template>
    <div class="flex h-max w-max items-center gap-2">
        <div class="cb-wrap">
            <input
                id="morning"
                type="checkbox"
                :class="{ 'inp-cbx': true }"
                :checked="checked"
                @change="(e) => $emit('changed', (e.target as HTMLInputElement).checked)"
            />
            <label
                class="cbx"
                for="morning"
            >
                <span>
                    <svg
                        width="12px"
                        height="10px"
                    >
                        <use xlink:href="#check-4"></use>
                    </svg>
                </span>
                <span v-if="label">{{ label }} </span>
            </label>
            <svg class="inline-svg">
                <symbol
                    id="check-4"
                    viewbox="0 0 12 10"
                >
                    <polyline points="1.5 6 4.5 9 10.5 1"></polyline>
                </symbol>
            </svg>
        </div>
    </div>
</template>

<script setup lang="ts">
defineProps<{ label?: string; checked: boolean }>()
defineEmits<{
    (e: 'changed', checked: boolean): void
}>()
</script>

<style scoped>
.cb-wrap .cbx {
    -webkit-user-select: none;
    user-select: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 6px;
    overflow: hidden;
    transition: all 200ms ease-in-out;
    display: inline-flex;
    text-wrap: nowrap;
    flex-wrap: nowrap;
}

.cb-wrap .cbx:not(:last-child) {
    margin-right: 6px;
}

.cb-wrap .cbx:hover {
    background: rgb(from var(--color-theme-primary) r g b / 15%);
}

.cb-wrap .cbx span {
    float: left;
    vertical-align: middle;
    transform: translate3d(0, 0, 0);
}

.cb-wrap .cbx span:first-child {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 4px;
    transform: scale(1);
    border: 1px solid var(--color-border-primary);
    transition: all 200ms ease-in-out;
}

.cb-wrap .cbx span:first-child svg {
    position: absolute;
    top: 3px;
    left: 2px;
    fill: none;
    stroke: #fff;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-dasharray: 16px;
    stroke-dashoffset: 16px;
    transition: all 100ms ease-in-out;
    transform: translate3d(0, 0, 0);
}

.cb-wrap .cbx span:last-child {
    padding-left: 8px;
    line-height: 18px;
}

.cb-wrap .cbx:hover span:first-child {
    border-color: var(--color-theme-primary);
}

.cb-wrap .inp-cbx {
    position: absolute;
    visibility: hidden;
}

.cb-wrap .inp-cbx:checked + .cbx span:first-child {
    background: var(--color-theme-primary);
    border-color: var(--color-theme-primary);
    animation: wave 150ms ease-in-out;
}

.cb-wrap .inp-cbx:checked + .cbx span:first-child svg {
    stroke-dashoffset: 0;
}

.cb-wrap .inline-svg {
    position: absolute;
    width: 0;
    height: 0;
    pointer-events: none;
    user-select: none;
}

@media screen and (max-width: 640px) {
    .cb-wrap .cbx {
        width: 100%;
        display: inline-block;
    }
}

@-moz-keyframes wave {
    50% {
        transform: scale(0.9);
    }
}

@-webkit-keyframes wave {
    50% {
        transform: scale(0.9);
    }
}

@-o-keyframes wave {
    50% {
        transform: scale(0.9);
    }
}

@keyframes wave {
    50% {
        transform: scale(0.9);
    }
}
</style>
