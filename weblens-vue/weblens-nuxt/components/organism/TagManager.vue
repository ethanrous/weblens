<template>
    <Teleport to="body">
        <div
            :class="{
                'fullscreen-modal p-12 transition lg:p-48': true,
                'pointer-events-none opacity-0': !visible,
            }"
        >
            <div
                ref="modal"
                class="bg-background-primary flex h-full w-full flex-col gap-3 rounded border p-4"
                @click.stop
            >
                <div class="flex items-center gap-2">
                    <IconTags size="20" />
                    <h4>Tags</h4>
                    <WeblensButton
                        :class="{ 'ml-auto': true }"
                        :square-size="32"
                        type="outline"
                        @click="emit('close')"
                    >
                        <IconX size="18" />
                    </WeblensButton>
                </div>

                <div
                    v-if="tagsStore.tagsList.length > 0"
                    class="flex flex-1 flex-col gap-0.5 overflow-y-auto"
                >
                    <div
                        v-for="tag in tagsStore.tagsList"
                        :key="tag.id"
                        class="group flex items-center gap-2 rounded px-2 py-1.5"
                    >
                        <template v-if="editingTagID === tag.id">
                            <WeblensInput
                                v-model:value="editName"
                                auto-focus
                                :class="{ 'h-8! flex-1': true }"
                                @submit="saveEdit(tag.id!)"
                            />
                            <div class="flex gap-0.5">
                                <div
                                    v-for="color in presetColors"
                                    :key="color"
                                    :class="{
                                        'h-4 w-4 cursor-pointer rounded-full border transition': true,
                                        'border-text-primary': editColor === color,
                                        'border-transparent': editColor !== color,
                                    }"
                                    :style="{ backgroundColor: color }"
                                    @click="editColor = color"
                                />
                            </div>
                            <WeblensButton
                                :square-size="28"
                                @click="saveEdit(tag.id!)"
                            >
                                <IconCheck size="14" />
                            </WeblensButton>
                            <WeblensButton
                                :square-size="28"
                                type="outline"
                                @click="editingTagID = ''"
                            >
                                <IconX size="14" />
                            </WeblensButton>
                        </template>

                        <template v-else>
                            <span
                                class="h-3 w-3 shrink-0 rounded-full"
                                :style="{ backgroundColor: tag.color }"
                            />
                            <span class="flex-1 truncate">{{ tag.name }}</span>
                            <span class="text-text-tertiary text-xs">
                                {{ (tag.fileIDs ?? []).length }}
                                {{ (tag.fileIDs ?? []).length === 1 ? 'file' : 'files' }}
                            </span>
                            <span
                                class="clickable text-text-tertiary opacity-0 group-hover:opacity-100"
                                title="View files"
                                @click="viewTagFiles(tag.id!)"
                            >
                                <IconEye size="14" />
                            </span>
                            <span
                                class="clickable text-text-tertiary opacity-0 group-hover:opacity-100"
                                @click="startEdit(tag)"
                            >
                                <IconPencil size="14" />
                            </span>
                            <span
                                class="clickable text-text-tertiary opacity-0 group-hover:opacity-100"
                                @click="handleDelete(tag.id!)"
                            >
                                <IconTrash size="14" />
                            </span>
                        </template>
                    </div>
                </div>

                <span
                    v-else
                    class="text-text-tertiary flex-1 py-6 text-center"
                >
                    No tags yet. Create your first tag below.
                </span>

                <div class="border-border-primary border-t pt-2">
                    <div
                        v-if="creating"
                        class="flex flex-col gap-2"
                    >
                        <WeblensInput
                            v-model:value="newName"
                            placeholder="Tag name"
                            auto-focus
                            @submit="handleCreate"
                        />
                        <div class="flex flex-wrap gap-1.5">
                            <div
                                v-for="color in presetColors"
                                :key="color"
                                :class="{
                                    'h-6 w-6 cursor-pointer rounded-full border-2 transition': true,
                                    'border-text-primary': newColor === color,
                                    'hover:border-text-tertiary border-transparent': newColor !== color,
                                }"
                                :style="{ backgroundColor: color }"
                                @click="newColor = color"
                            />
                        </div>
                        <div class="flex gap-2">
                            <WeblensButton
                                label="Create"
                                fill-width
                                :disabled="!newName.trim()"
                                @click="handleCreate"
                            >
                                <IconPlus size="16" />
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
                        @click="startCreate"
                    >
                        <IconPlus size="16" />
                    </WeblensButton>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { IconCheck, IconEye, IconPencil, IconPlus, IconTags, IconTrash, IconX } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import WeblensInput from '../atom/WeblensInput.vue'
import useTagsStore from '~/stores/tags'
import type { TagInfo } from '~/stores/tags'
import { onClickOutside, onKeyDown } from '@vueuse/core'

const tagsStore = useTagsStore()

const props = defineProps<{
    visible: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
}>()

const modal = useTemplateRef('modal')

onClickOutside(modal, () => {
    if (props.visible) {
        emit('close')
    }
})

onKeyDown(['Escape'], () => {
    if (props.visible) {
        emit('close')
    }
})

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

const creating = ref(false)
const newName = ref('')
const newColor = ref(presetColors[0])

const editingTagID = ref('')
const editName = ref('')
const editColor = ref('')

function startCreate() {
    creating.value = true
    newName.value = ''
    newColor.value = presetColors[Math.floor(Math.random() * presetColors.length)]
}

async function handleCreate() {
    const name = newName.value.trim()
    if (!name) return
    await tagsStore.createTag(name, newColor.value)
    creating.value = false
}

function startEdit(tag: TagInfo) {
    editingTagID.value = tag.id ?? ''
    editName.value = tag.name ?? ''
    editColor.value = tag.color ?? ''
}

async function saveEdit(tagID: string) {
    const name = editName.value.trim()
    if (!name) return
    await tagsStore.updateTag(tagID, name, editColor.value)
    editingTagID.value = ''
}

function viewTagFiles(tagID: string) {
    emit('close')
    navigateTo(`/files/tag/${tagID}`)
}

async function handleDelete(tagID: string) {
    await tagsStore.deleteTag(tagID)
}

onMounted(() => {
    if (tagsStore.tagsList.length === 0) {
        tagsStore.fetchTags()
    }
})
</script>
