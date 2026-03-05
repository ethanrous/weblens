<template>
    <div
        ref="selectorRef"
        class="bg-background-primary border-border-primary flex w-56 flex-col gap-1 rounded border p-2 shadow-lg"
    >
        <span class="text-text-secondary px-1 text-xs font-semibold uppercase">Tags</span>

        <div
            v-if="tagsStore.tagsList.length > 0"
            class="flex max-h-48 flex-col overflow-y-auto"
        >
            <div
                v-for="tag in tagsStore.tagsList"
                :key="tag.id"
                class="clickable flex items-center gap-2 px-1"
                @click="toggleTag(tag.id)"
            >
                <span
                    class="h-2.5 w-2.5 shrink-0 rounded-full"
                    :style="{ backgroundColor: tag.color }"
                />
                <span class="truncate text-sm">{{ tag.name }}</span>
                <IconCheck
                    v-if="isTagApplied(tag.id)"
                    size="14"
                    class="text-text-secondary ml-auto shrink-0"
                />
                <IconMinus
                    v-else-if="isTagMixed(tag.id)"
                    size="14"
                    class="text-text-tertiary ml-auto shrink-0"
                />
            </div>
        </div>

        <span
            v-else
            class="text-text-tertiary px-1 py-2 text-center text-sm"
        >
            No tags yet
        </span>

        <div class="border-border-primary mt-1 border-t pt-1.5">
            <div
                v-if="creating"
                class="flex flex-col gap-1.5"
            >
                <WeblensInput
                    v-model:value="newTagName"
                    placeholder="Tag name"
                    auto-focus
                    @submit="handleCreate"
                />
                <div class="flex flex-wrap gap-1 px-0.5">
                    <div
                        v-for="color in presetColors"
                        :key="color"
                        :class="{
                            'h-5 w-5 cursor-pointer rounded-full border-2 transition': true,
                            'border-text-primary': selectedColor === color,
                            'hover:border-text-tertiary border-transparent': selectedColor !== color,
                        }"
                        :style="{ backgroundColor: color }"
                        @click="selectedColor = color"
                    />
                </div>
                <div class="flex gap-1">
                    <WeblensButton
                        label="Create"
                        fill-width
                        :disabled="!newTagName.trim()"
                        @click="handleCreate"
                    >
                        <IconPlus size="14" />
                    </WeblensButton>
                    <WeblensButton
                        label="Cancel"
                        type="outline"
                        @click="creating = false"
                    />
                </div>
            </div>

            <WeblensButton
                v-else
                label="New Tag"
                type="outline"
                fill-width
                @click="startCreating"
            >
                <IconPlus size="14" />
            </WeblensButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconCheck, IconMinus, IconPlus } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import WeblensInput from '../atom/WeblensInput.vue'
import useTagsStore from '~/stores/tags'

const tagsStore = useTagsStore()

const props = defineProps<{
    fileIDs: string[]
}>()

const creating = ref(false)
const newTagName = ref('')

const presetColors = [
    '#e74c3c',
    '#e67e22',
    '#f1c40f',
    '#2ecc71',
    '#1abc9c',
    '#3498db',
    '#9b59b6',
    '#e91e63',
    '#795548',
    '#607d8b',
    '#34495e',
    '#7f8c8d',
]

const selectedColor = ref(presetColors[0])

function isTagApplied(tagID: string): boolean {
    const tag = tagsStore.tags.get(tagID)
    if (!tag) return false
    return props.fileIDs.every((fID) => tag.fileIDs.includes(fID))
}

function isTagMixed(tagID: string): boolean {
    const tag = tagsStore.tags.get(tagID)
    if (!tag) return false
    const count = props.fileIDs.filter((fID) => tag.fileIDs.includes(fID)).length
    return count > 0 && count < props.fileIDs.length
}

async function toggleTag(tagID: string) {
    if (isTagApplied(tagID)) {
        await tagsStore.removeFilesFromTag(tagID, props.fileIDs)
    } else {
        await tagsStore.addFilesToTag(tagID, props.fileIDs)
    }
}

function startCreating() {
    creating.value = true
    newTagName.value = ''
    selectedColor.value = presetColors[Math.floor(Math.random() * presetColors.length)]
}

async function handleCreate() {
    const name = newTagName.value.trim()
    if (!name) return

    const tag = await tagsStore.createTag(name, selectedColor.value)
    creating.value = false

    if (props.fileIDs.length > 0) {
        await tagsStore.addFilesToTag(tag.id, props.fileIDs)
    }
}

onMounted(() => {
    if (tagsStore.tagsList.length === 0) {
        tagsStore.fetchTags()
    }
})
</script>
