<template>
    <div
        ref="selectDropdown"
        :class="{
            'relative z-60 flex h-full w-full max-w-32 min-w-10 transition-[z-index]': true,
            'z-90': isOpen,
            'delay-300': !isOpen,
        }"
        @click="isOpen = !isOpen"
    >
        <div
            ref="innerRef"
            :class="{
                'border-theme-primary absolute flex min-h-full w-full flex-col justify-start overflow-hidden rounded border transition-[height,width,scale]': true,
                'before:bg-background-primary/80 backdrop-blur-sm before:absolute before:z-[-1] before:h-full before:w-full': true,
                'rounded-r-none border-r-transparent': merge === 'right' && !isOpen,
                'z-90 h-max w-32! shadow-lg': isOpen,
            }"
            :style="{
                height: isOpen ? (opts.length + 1) * (selectSize.height.value + 2) - 2 + 'px' : 'calc(100% - 2px)',
            }"
        >
            <div
                :class="{
                    'bg-theme-primary/35 absolute w-full rounded transition-[top]': true,
                    'rounded-r-none': merge === 'right' && !isOpen,
                }"
                :style="{
                    height: selectSize.height.value - 2 + 'px',
                    top: isOpen ? optionHoverIndex * (selectSize.height.value + 2) + 'px' : '0px',
                }"
            />
            <div
                :class="{
                    'absolute flex h-max w-full flex-col gap-1 px-2': true,
                    'p-0!': iconOnly && !isOpen,
                }"
            >
                <div
                    :class="{
                        option: true,
                        'w-full': !isOpen,
                        'justify-center': iconOnly && !isOpen,
                    }"
                    :style="{
                        height: selectSize.height.value - 2 + 'px',
                    }"
                    @mouseover="optionHoverIndex = 0"
                >
                    <component
                        :is="selectedOption?.icon"
                        v-if="selectedOption?.icon"
                        :class="{
                            'transition-[margin]': true,
                        }"
                    />
                    <span
                        v-if="!iconOnly || isOpen"
                        :class="{ 'ml-1': true }"
                    >
                        {{ selectedOption?.label }}
                    </span>
                </div>
                <div
                    v-for="(option, index) in opts"
                    :key="option.value"
                >
                    <div
                        :class="{
                            'option cursor-pointer': true,
                            'pointer-events-none': option.disabled,
                        }"
                        :style="{
                            height: selectSize.height.value - 2 + 'px',
                        }"
                        @mouseover="optionHoverIndex = index + 1"
                        @click="value = option.value"
                    >
                        <component
                            :is="option?.icon"
                            v-if="option?.icon"
                            :class="{
                                'mr-1 transition-[margin]': true,
                                'text-text-tertiary': option.disabled,
                            }"
                        />
                        <span :class="{ 'text-text-tertiary': option.disabled }">
                            {{ option?.label }}
                        </span>
                    </div>
                </div>
            </div>

            <span
                ref="textSizeRef"
                :class="{ 'pointer-events-none absolute opacity-0': true }"
            >
                {{ selectedOption?.label }}
            </span>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { Icon } from '@tabler/icons-vue'
import { onClickOutside, useElementSize } from '@vueuse/core'

const value = defineModel<keyof typeof props.options>('value')

const isOpen = ref<boolean>(false)
const optionHoverIndex = ref<number>(1)

const selectDropdown = ref<HTMLDivElement>()
const innerRef = ref<HTMLDivElement>()
const textSizeRef = ref<HTMLSpanElement>()

const selectSize = useElementSize(selectDropdown)
const innerSize = useElementSize(innerRef)
const textSize = useElementSize(textSizeRef)
onClickOutside(selectDropdown, () => {
    isOpen.value = false
})

export type SelectOption = {
    label: string
    value?: string
    icon?: Icon
    disabled?: boolean
}

const props = defineProps<{
    options: Record<string, SelectOption>
    merge?: 'right' | 'left'
}>()

const opts = computed(() => {
    const opts = { ...props.options }

    return Object.entries(opts)
        .filter(([v]) => v !== value.value)
        .map(([k, v]) => {
            if (!v.value) {
                v.value = k
            }

            return v
        })
})

const selectedOption = computed(() => {
    if (!value.value || !props.options[value.value]) {
        return { label: 'Select an option', value: '' }
    }

    return props.options[value.value]
})

const iconOnly = computed(() => {
    return innerSize.width.value <= 40 + textSize.width.value
})
</script>

<style scoped lang="less">
.option {
    display: flex;
    align-items: center;
    gap: var(--wl-gap-sm);
    color: var(--color-text-primary);
    user-select: none;
    width: 100%;
}
</style>
