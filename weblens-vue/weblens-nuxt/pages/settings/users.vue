<template>
    <div :class="{ 'flex flex-col gap-3': true }">
        <h4>Users</h4>
        <Table
            :columns="['username', 'role', 'online', 'activated', 'delete']"
            :rows="rows"
        >
            <template
                v-for="(row, index) in rows"
                :key="index"
                #[`online-${index}`]
            >
                <div :class="{ 'flex w-full justify-center': true }">
                    <WebsocketStatus :status="(row['online'] as TableTypes[TableType.Slot]).data ? 'OPEN' : 'CLOSED'" />
                </div>
            </template>
        </Table>
        <div :class="{ 'mt-4 flex flex-col gap-3': true }">
            <h4>Add User</h4>
            <div :class="{ 'flex w-72 flex-col gap-2 rounded border p-4': true }">
                <WeblensInput
                    v-model="newUsername"
                    placeholder="Username"
                />
                <WeblensInput
                    v-model="newPassword"
                    placeholder="Password"
                    password
                />
                <WeblensButton
                    label="Create User"
                    :class="{ 'my-2': true }"
                    :disabled="!newUsername || !newPassword"
                    @click.stop="handleNewUser"
                >
                    <IconUserPlus />
                </WeblensButton>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconTrash, IconUserPlus } from '@tabler/icons-vue'
import { useWeblensAPI } from '~/api/AllApi'
import Table from '~/components/atom/Table.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import WebsocketStatus from '~/components/molecule/WebsocketStatus.vue'
import { TableType, type TableColumn, type TableColumns, type TableTypes } from '~/types/table'
import User, { UserPermissions } from '~/types/user'

const userStore = useUserStore()

const newUsername = ref('')
const newPassword = ref('')

const { data: users, refresh } = useAsyncData('users', async () => {
    const { data } = await useWeblensAPI().UsersAPI.getUsers()
    return data
})

const rows = computed<TableColumns>(() => {
    if (!users.value) {
        return []
    }

    return users.value.map((user, index) => {
        let activated: TableColumn
        if (!user.activated) {
            activated = {
                tableType: TableType.Button,
                label: 'Activate',
                onclick: async () => {
                    await useWeblensAPI().UsersAPI.activateUser(user.username, true)
                    refresh()
                },
            }
        } else {
            activated = 'Activated'
        }

        const thing: Record<string, TableColumn> = {
            username: user.username,
            role: User.GetPermissionLevelName(user.permissionLevel),
            activated: activated,
            online: { tableType: TableType.Slot, key: `online-${index}`, data: user.isOnline },
            delete: {
                tableType: TableType.Button,
                flavor: 'danger',
                disabled:
                    (user.permissionLevel >= UserPermissions.Admin &&
                        userStore.user.GetPermissionLevel() < UserPermissions.Admin) ||
                    user.username === userStore.user.username,
                icon: IconTrash,
                onclick: async () => {
                    await useWeblensAPI().UsersAPI.deleteUser(user.username)
                    refresh()
                },
            },
        }

        return thing
    })
})

async function handleNewUser() {
    if (!newUsername.value || !newPassword.value) {
        return
    }

    await useWeblensAPI().UsersAPI.createUser({
        username: newUsername.value,
        password: newPassword.value,
        autoActivate: true,
        fullName: '',
    })

    newUsername.value = ''
    newPassword.value = ''
    await refresh()
}
</script>
