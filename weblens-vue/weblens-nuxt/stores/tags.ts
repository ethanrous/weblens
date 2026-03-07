import { defineStore } from 'pinia'
import type { GithubComEthanrousWeblensModelsTagTag } from '@ethanrous/weblens-api'
import { useWeblensAPI } from '~/api/AllApi'

export type TagInfo = GithubComEthanrousWeblensModelsTagTag

const useTagsStore = defineStore('tags', () => {
    const tags = ref<Map<string, TagInfo>>(new Map())
    const loading = ref(false)

    const tagsList = computed(() => {
        return [...tags.value.values()].sort((a, b) => (a.name ?? '').localeCompare(b.name ?? ''))
    })

    function getTagsByFileID(fileID: string): TagInfo[] {
        return [...tags.value.values()].filter((tag) => tag.fileIDs?.includes(fileID))
    }

    async function fetchTags() {
        loading.value = true
        try {
            const { data } = await useWeblensAPI().TagsAPI.getUserTags()
            const newTags = new Map<string, TagInfo>()
            for (const tag of data) {
                newTags.set(tag.id!, tag)
            }
            tags.value = newTags
        } catch (err) {
            console.error('Failed to fetch tags:', err)
        } finally {
            loading.value = false
        }
    }

    async function createTag(name: string, color: string) {
        const { data: tag } = await useWeblensAPI().TagsAPI.createTag({ name, color })
        const newTags = new Map(tags.value)
        newTags.set(tag.id!, tag)
        tags.value = newTags
        return tag
    }

    async function updateTag(tagID: string, name?: string, color?: string) {
        await useWeblensAPI().TagsAPI.updateTag(tagID, { name, color })
        const existing = tags.value.get(tagID)
        if (existing) {
            const newTags = new Map(tags.value)
            newTags.set(tagID, {
                ...existing,
                name: name ?? existing.name,
                color: color ?? existing.color,
                updated: new Date().toISOString(),
            })
            tags.value = newTags
        }
    }

    async function deleteTag(tagID: string) {
        await useWeblensAPI().TagsAPI.deleteTag(tagID)
        const newTags = new Map(tags.value)
        newTags.delete(tagID)
        tags.value = newTags
    }

    async function addFilesToTag(tagID: string, fileIDs: string[]) {
        await useWeblensAPI().TagsAPI.addFilesToTag(tagID, { fileIDs })
        const existing = tags.value.get(tagID)
        if (existing) {
            const newFileIDs = new Set(existing.fileIDs ?? [])
            for (const fID of fileIDs) {
                newFileIDs.add(fID)
            }
            const newTags = new Map(tags.value)
            newTags.set(tagID, { ...existing, fileIDs: [...newFileIDs] })
            tags.value = newTags
        }
    }

    async function removeFilesFromTag(tagID: string, fileIDs: string[]) {
        await useWeblensAPI().TagsAPI.removeFilesFromTag(tagID, { fileIDs })
        const existing = tags.value.get(tagID)
        if (existing) {
            const removeSet = new Set(fileIDs)
            const newTags = new Map(tags.value)
            newTags.set(tagID, {
                ...existing,
                fileIDs: (existing.fileIDs ?? []).filter((fID) => !removeSet.has(fID)),
            })
            tags.value = newTags
        }
    }

    function removeFileFromLocalTags(fileID: string) {
        const newTags = new Map<string, TagInfo>()
        for (const [id, tag] of tags.value) {
            if (tag.fileIDs?.includes(fileID)) {
                newTags.set(id, { ...tag, fileIDs: tag.fileIDs.filter((fID) => fID !== fileID) })
            } else {
                newTags.set(id, tag)
            }
        }
        tags.value = newTags
    }

    return {
        tags,
        loading,
        tagsList,

        getTagsByFileID,
        fetchTags,
        createTag,
        updateTag,
        deleteTag,
        addFilesToTag,
        removeFilesFromTag,
        removeFileFromLocalTags,
    }
})

export default useTagsStore
