<template>
    <div
        :class="{
            'fullscreen-modal p-12 transition xl:p-48': true,
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

            <div :class="{ 'z-99 flex w-full items-center gap-2': true }">
                <UserSearch
                    :class="{ 'min-w-10': true }"
                    :exclude-fn="excludeFn"
                    @select:user="addAccessor"
                />
                <WeblensButton
                    :class="{ 'min-w-0 sm:min-w-40': true }"
                    :label="share?.timelineOnly ? 'Timeline Only' : 'Timeline + Files'"
                    :type="share?.timelineOnly ? 'outline' : 'default'"
                    allow-collapse
                    @click="toggleTimelienOnly"
                >
                    <IconFileOff v-if="share?.timelineOnly" />
                    <IconFile v-else />
                </WeblensButton>
            </div>
            <Table
                :columns="['username', 'canDownload', 'canEdit', 'canDelete', 'unshare']"
                :rows="rows"
            />
            <span :class="{ 'text-text-secondary mt-4': true }">Public Share Settings</span>
            <Table
                :columns="['public', 'canViewFiles', 'canDownload', 'canEdit', 'canDelete']"
                :rows="publicShareRows"
            />
            <CopyBox
                :text="share?.ID() ? share?.GetLink() : undefined"
                :class="{ 'mt-auto': true }"
            />
            <div :class="{ 'flex gap-2': true }">
                <WeblensButton
                    label="Revoke Share"
                    :flavor="'danger'"
                    :disabled="!share?.ID()"
                    @click.stop="deleteShare()"
                >
                    <IconUserOff />
                </WeblensButton>
                <WeblensButton
                    label="Done"
                    :class="{ 'ml-auto w-1/4': true }"
                    @click.stop="menuStore.setSharing(false)"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconFile, IconFileOff, IconLock, IconLockOpen, IconUserOff } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import type WeblensFile from '~/types/weblensFile'
import FileIcon from '../atom/FileIcon.vue'
import UserSearch from '../molecule/UserSearch.vue'
import CopyBox from '../molecule/CopyBox.vue'
import { onClickOutside } from '@vueuse/core'
import Table from '../atom/Table.vue'
import type { UserInfo } from '@ethanrous/weblens-api'
import { TableType, type TableColumn, type TableColumns, type TableRow } from '~/types/table'
import { useWeblensAPI } from '~/api/AllApi'
import { UNAUTHENTICATED_USER_NAME } from '~/types/user'

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
            icon: IconUserOff,
            onclick: async () => {
                if (!share.value) return
                await share.value.removeAccessor(u.username)
                share.value = share.value.clone()
            },
        },
    }))
})

const rows = computed<TableRow[]>(() => {
    let rows: TableRow[] = [...accessors.value]
    if (rows.length === 0) {
        rows = [
            {
                sectionHeader: true,
                text: 'Not shared with anyone',
                icon: IconUserOff,
                className: 'text-text-tertiary',
            },
        ]
    }

    return rows
})

const publicShareRows = computed<TableRow[]>(() => {
    if (!share.value) {
        return []
    }

    return [
        {
            public: {
                tableType: TableType.Button,
                checked: share.value.IsPublic(),
                label: share.value.IsPublic() ? 'Public' : 'Private',
                type: share.value.IsPublic() ? 'default' : 'outline',
                icon: share.value.IsPublic() ? IconLockOpen : IconLock,
                onClick: async () => {
                    await toggleIsPublic()
                },
            },
            canViewFiles: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions[UNAUTHENTICATED_USER_NAME].canView ?? false,
                disabled: !share.value.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions[UNAUTHENTICATED_USER_NAME],
                        canView: c,
                    })

                    share.value = share.value.clone()
                },
            },
            canDownload: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions[UNAUTHENTICATED_USER_NAME].canDownload ?? false,
                disabled: !share.value.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions[UNAUTHENTICATED_USER_NAME],
                        canDownload: c,
                    })

                    share.value = share.value.clone()
                },
            },
            canEdit: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions[UNAUTHENTICATED_USER_NAME].canEdit ?? false,
                disabled: !share.value.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions[UNAUTHENTICATED_USER_NAME],
                        canEdit: c,
                    })

                    share.value = share.value.clone()
                },
            },
            canDelete: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions[UNAUTHENTICATED_USER_NAME].canDelete ?? false,
                disabled: !share.value.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions[UNAUTHENTICATED_USER_NAME],
                        canDelete: c,
                    })

                    share.value = share.value.clone()
                },
            },
        },
    ]
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

async function deleteShare() {
    if (!share.value) {
        console.error('No share to delete')
        return
    }

    try {
        await useWeblensAPI().SharesAPI.deleteFileShare(share.value.ID())
        share.value = null
        menuStore.setSharing(false)
    } catch (e) {
        console.error('Failed to delete share', e)
    }
}
</script>
