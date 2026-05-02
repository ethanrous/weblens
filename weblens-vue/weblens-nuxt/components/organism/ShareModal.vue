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
                'bg-background-primary flex h-full w-full flex-col gap-4 overflow-y-auto rounded border p-4': true,
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

            <Table
                :columns="['username', 'canDownload', 'canEdit', 'canDelete', 'unshare']"
                :rows="rows"
            />
            <UserSearch
                :class="{ 'min-w-10': true }"
                :exclude-fn="excludeFn"
                @select:user="addAccessor"
            />

            <span :class="{ 'text-text-secondary mt-4': true }">Public Share Settings</span>
            <Table
                :columns="['public', 'canViewFiles', 'canDownload', 'canEdit']"
                :rows="publicShareRows"
            />

            <div :class="{ 'mt-auto inline-flex w-full items-center gap-2': true }">
                <WeblensButton @click="doTimeline = !doTimeline">
                    <IconPhoto v-if="doTimeline" />
                    <IconFolder v-if="!doTimeline" />
                </WeblensButton>

                <CopyBox
                    :text="share?.ID() ? share?.GetLink(doTimeline) : undefined"
                    :class="{ grow: true }"
                />
            </div>

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
import { IconFolder, IconLock, IconLockOpen, IconPhoto, IconUserOff } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import type WeblensFile from '~/types/weblensFile'
import FileIcon from '../atom/FileIcon.vue'
import UserSearch from '../molecule/UserSearch.vue'
import CopyBox from '../molecule/CopyBox.vue'
import { onClickOutside } from '@vueuse/core'
import Table from '../atom/Table.vue'
import type { UserInfo } from '@ethanrous/weblens-api'
import { TableType, type TableColumn, type TableColumns, type TableRow } from '~/types/table'
import { UNAUTHENTICATED_USER_NAME } from '~/types/user'

const menuStore = useContextMenuStore()

const doTimeline = ref<boolean>(false)

const modal = ref<HTMLDivElement>()
onClickOutside(modal, () => {
    if (menuStore.isSharing) menuStore.setSharing(false)
})

const props = defineProps<{
    file: WeblensFile
}>()

const { data: share, refresh: refreshShare } = useAsyncData(
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

                await refreshShare()
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

                await refreshShare()
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

                await refreshShare()
            },
        },
        unshare: {
            flavor: 'danger',
            tableType: TableType.Button,
            icon: IconUserOff,
            onClick: async () => {
                if (!share.value) return
                await share.value.removeAccessor(u.username)

                await refreshShare()
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
    return [
        {
            public: {
                tableType: TableType.Button,
                label: share.value?.IsPublic() ? 'Public' : 'Private',
                type: share.value?.IsPublic() ? 'default' : 'outline',
                icon: share.value?.IsPublic() ? IconLockOpen : IconLock,
                onClick: async () => {
                    let sh = share.value
                    if (!sh) {
                        sh = await props.file.Share()
                    }

                    await sh.toggleIsPublic()
                    await refreshShare()
                },
            },
            canViewFiles: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions?.[UNAUTHENTICATED_USER_NAME]?.canView ?? false,
                disabled: !share.value?.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions?.[UNAUTHENTICATED_USER_NAME],
                        canView: c,
                    })

                    await refreshShare()
                },
            },
            canDownload: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions?.[UNAUTHENTICATED_USER_NAME]?.canDownload ?? false,
                disabled: !share.value?.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions?.[UNAUTHENTICATED_USER_NAME],
                        canDownload: c,
                    })

                    await refreshShare()
                },
            },
            canEdit: {
                tableType: TableType.Checkbox,
                checked: share.value?.permissions?.[UNAUTHENTICATED_USER_NAME]?.canEdit ?? false,
                disabled: !share.value?.IsPublic(),
                onchanged: async (c: boolean) => {
                    if (!share.value) return
                    await share.value.updateAccessorPerms(UNAUTHENTICATED_USER_NAME, {
                        ...share.value?.permissions?.[UNAUTHENTICATED_USER_NAME],
                        canEdit: c,
                    })

                    await refreshShare()
                },
            },
        },
    ]
})

async function addAccessor(user: UserInfo) {
    let sh = share.value
    if (!sh) {
        sh = await props.file.Share()
    }

    await sh.addAccessor(user.username)
    await refreshShare()
}

function excludeFn(u: UserInfo) {
    if (share.value?.accessors.map((u) => u.username).includes(u.username)) {
        return false
    }

    return true
}

async function deleteShare() {
    if (!share.value) {
        console.error('No share to delete')
        return
    }

    await props.file.DeleteShare()
    await refreshShare()
}
</script>
