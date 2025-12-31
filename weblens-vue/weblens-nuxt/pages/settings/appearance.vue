<template>
    <div>
        <h1 class="text-2xl font-bold">Appearance Settings</h1>
        <p class="mt-2 text-gray-600">Customize the look and feel of your Weblens experience.</p>
        <div class="mt-4">
            <h2 class="text-xl font-semibold">Theme</h2>
            <p class="mt-1 text-gray-500">Choose your preferred theme.</p>
            <div :class="{ 'flex gap-2': true }">
                <span>Dark</span>
                <input
                    type="checkbox"
                    :checked="isDark"
                    @change="(e) => (isDark = (e.target as HTMLInputElement).checked)"
                />
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
import { useDark } from '@vueuse/core'

const isDark = useDark({
    selector: 'body',
    attribute: 'color-scheme',
    valueDark: 'dark',
    valueLight: 'light',
    onChanged: (isDark) => {
        document?.documentElement?.style?.setProperty('color-scheme', isDark ? 'dark' : 'light')
    },
})
</script>
