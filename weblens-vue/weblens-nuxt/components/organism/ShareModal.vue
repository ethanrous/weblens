<template>
    <div
        :class="{
            'fullscreen-modal p-12 transition lg:p-48': true,
            'pointer-events-none opacity-0': !menuStore.isSharing,
        }"
    >
        <div
            ref="modal"
            :class="{
                'bg-background-primary flex h-full w-full flex-col gap-4 rounded border p-4': true,
            }"
            @click.stop
        >
            <div :class="{ 'flex items-center gap-1': true }">
                <h4>Share</h4>
                <FileIcon
                    :file="file"
                    :with-name="true"
                />
            </div>

            <div :class="{ 'z-99 flex w-full gap-2': true }">
                <UserSearch
                    :exclude-fn="excludeFn"
                    @select:user="addAccessor"
                />
                <WeblensButton
                    :class="{ 'w-10 sm:w-40': true }"
                    :label="share?.IsPublic() ? 'Public' : 'Private'"
                    :type="share?.IsPublic() ? 'default' : 'outline'"
                    allow-collapse
                    @click="toggleIsPublic"
                >
                    <IconLock v-if="!share?.IsPublic()" />
                    <IconLockOpen v-else />
                </WeblensButton>
                <WeblensButton
                    :class="{ 'w-10 sm:w-40': true }"
                    label="Timeline Only"
                    :type="share?.timelineOnly ? 'default' : 'outline'"
                    allow-collapse
                    @click="toggleTimelienOnly"
                >
                    <IconPhoto v-if="share?.timelineOnly" />
                    <IconPhotoOff v-else />
                </WeblensButton>
            </div>
            <Table
                :columns="['username', 'canDownload', 'canEdit', 'canDelete', 'unshare']"
                :rows="accessors"
            />
            <CopyBox
                :text="share?.ID() ? share?.GetLink() : undefined"
                :class="{ 'mt-auto': true }"
            />
            <div :class="{ 'flex gap-2': true }">
                <WeblensButton
                    label="Done"
                    :class="{ 'ml-auto w-1/2': true }"
                    @click.stop="menuStore.setSharing(false)"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconLock, IconLockOpen, IconPhoto, IconPhotoOff, IconTrash } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import type WeblensFile from '~/types/weblensFile'
import FileIcon from '../atom/FileIcon.vue'
import UserSearch from '../molecule/UserSearch.vue'
import CopyBox from '../molecule/CopyBox.vue'
import { onClickOutside } from '@vueuse/core'
import Table from '../atom/Table.vue'
import type { UserInfo } from '@ethanrous/weblens-api'
import { TableType, type TableColumn, type TableColumns } from '~/types/table'

const menuStore = useContextMenuStore()

const modal = ref<HTMLDivElement>()
onClickOutside(modal, () => {
    if (menuStore.isSharing) menuStore.setSharing(false)
})

const props = defineProps<{
    file: WeblensFile
}>()

const { data: share } = useAsyncData(
    'share-' + props.file.ID(),
    async () => {
        return await props.file.GetShare()
    },
    { deep: false, watch: [() => props.file.ID()] },
)

const accessors = computed<TableColumns>(() => {
    if (!share.value) {
        return []
    }

    return share.value.accessors.map<Record<string, TableColumn>>((u) => ({
        username: {
            tableType: TableType.Text,
            text: u.username,
        },
        canDownload: {
            tableType: TableType.Checkbox,
            checked: share.value?.permissions[u.username]?.canDownload ?? false,
            onchanged: async (c: boolean) => {
                if (!share.value) return
                await share.value.updateAccessorPerms(u.username, {
                    ...share.value?.permissions[u.username],
                    canDownload: c,
                })
                share.value = share.value.clone()
            },
        },
        canEdit: {
            tableType: TableType.Checkbox,
            checked: share.value?.permissions[u.username]?.canEdit ?? false,
            onchanged: async (c: boolean) => {
                if (!share.value) return
                await share.value.updateAccessorPerms(u.username, {
                    ...share.value?.permissions[u.username],
                    canEdit: c,
                })
                share.value = share.value.clone()
            },
        },
        canDelete: {
            tableType: TableType.Checkbox,
            checked: share.value?.permissions[u.username]?.canDelete ?? false,
            onchanged: async (c: boolean) => {
                if (!share.value) return
                await share.value.updateAccessorPerms(u.username, {
                    ...share.value?.permissions[u.username],
                    canDelete: c,
                })

                share.value = share.value.clone()
            },
        },
        unshare: {
            flavor: 'danger',
            tableType: TableType.Button,
            icon: IconTrash,
            onclick: async () => {
                if (!share.value) return
                await share.value.removeAccessor(u.username)
                share.value = share.value.clone()
            },
        },
    }))
})

async function addAccessor(user: UserInfo) {
    if (!share.value) {
        console.error('No share to add accessor')
        return
    }

    await share.value.addAccessor(user.username)
    share.value = share.value.clone()
}

function excludeFn(u: UserInfo) {
    if (share.value?.accessors.map((u) => u.username).includes(u.username)) {
        return false
    }

    return true
}

async function toggleIsPublic() {
    if (!share.value) {
        console.error('No share to toggle isPublic')
        return
    }

    await share.value.toggleIsPublic()
}

async function toggleTimelienOnly() {
    if (!share.value) {
        console.error('No share to toggle isPublic')
        return
    }

    await share.value.toggleTimelineOnly()
}
</script>
