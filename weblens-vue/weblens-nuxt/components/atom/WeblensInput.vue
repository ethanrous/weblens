<template>
    <div
        ref="inputContainer"
        :class="{
            'hover:bg-background-hover relative flex h-12 cursor-text items-center gap-2 rounded-lg border p-2 transition-[width]': true,
            'rounded-none first:rounded-l last:rounded-r': merge === 'row',
            'rounded-none first:rounded-t last:rounded-b': merge === 'column',
        }"
        @click="() => input?.focus()"
    >
        <slot />
        <input
            ref="input"
            v-model="value"
            :placeholder="placeholder"
            :type="type"
            :value="value"
            @input="
                (e) => {
                    emit('update:value', (e.target as HTMLInputElement).value)
                }
            "
            @keydown.escape="
                () => {
                    if (focused) {
                        input?.blur()
                    }
                }
            "
        />
        <div
            v-if="keyName"
            :class="{
                'text-text-tertiary pointer-events-none absolute right-2 p-1 text-nowrap transition': true,
                'opacity-0': focused || value,
            }"
        >
            <span>
                {{ keyName }}
            </span>
        </div>
        <div
            v-if="value && focused && showSubmit"
            :class="{
                'text-text-tertiary hover:text-text-primary hover:bg-card-background-secondary absolute top-1/2 right-2 -translate-y-1/2 rounded p-1 transition': true,
            }"
            @click.stop="
                () => {
                    emit('submit', value ?? '')
                    input?.blur()
                }
            "
        >
            <IconArrowRight />
        </div>
        <div
            v-if="clearButton && value && focused"
            :class="{
                'text-text-tertiary hover:text-text-primary hover:bg-card-background-secondary z-90 cursor-pointer rounded p-1 transition': true,
            }"
            @click.stop="handleClear"
        >
            <IconX />
        </div>
        <slot name="rightIcon" />
    </div>
</template>

<script setup lang="ts">
import { IconArrowRight, IconX } from '@tabler/icons-vue'
import { onKeyDown, useFocusWithin } from '@vueuse/core'

const inputContainer = ref<HTMLDivElement>()
const input = ref<HTMLInputElement>()
const { focused } = useFocusWithin(inputContainer)

const props = defineProps<{
    placeholder?: string
    password?: boolean
    keyName?: string
    buttonish?: boolean
    showSubmit?: boolean
    autoFocus?: boolean
    search?: boolean
    merge?: 'row' | 'column'
    clearButton?: boolean
}>()

const value = defineModel<string>('value')

const type = computed(() => {
    if (props.password) {
        return 'password'
    } else if (props.search) {
        return 'search'
    }

    return 'text'
})

const emit = defineEmits<{
    (e: 'update:value' | 'submit', value: string): void
    (e: 'clear' | 'focused'): void
}>()

function focus() {
    input.value?.focus()
}

function handleClear() {
    value.value = ''
    emit('clear')
    if (props.autoFocus) {
        input.value?.focus()
    } else {
        input.value?.blur()
    }
}

defineExpose({
    focus,
    inputContainer,
})

onKeyDown(['Enter'], (e) => {
    if (!focused.value) return
    e.preventDefault()
    emit('submit', value.value ?? '')
    input.value?.blur()
})

watchEffect(() => {
    if (focused.value) {
        if (props.search) {
            input.value?.select()
        }

        emit('focused')
    }
})

onMounted(() => {
    if (props.autoFocus) {
        // vueuse useFocusWithin won't see the focus event if we do it immediately on mounted,
        // so we wait for the next tick.
        nextTick().then(() => {
            input.value?.focus()
        })
    }
})
</script>
