<template>
    <div
        ref="inputContainerRef"
        :class="{ 'relative w-full': true }"
    >
        <WeblensInput
            v-model:value="search"
            placeholder="Find people..."
            :class="{ 'border-b-none rounded-b-none': users?.length && open }"
            clear-button
            @clear="search = ''"
            @escaped="open = false"
        />
        <div
            :class="{
                'bg-background-primary border-t-none absolute z-10 w-full overflow-hidden rounded-b border shadow transition': true,
                'pointer-events-none opacity-0': !users?.length || !open,
            }"
        >
            <div
                v-for="user in users"
                :key="user.username"
                :class="{ 'hover:bg-background-hover cursor-pointer p-2': true }"
                @click.stop.prevent="
                    () => {
                        emit('select:user', user)
                        open = false
                        search = ''
                    }
                "
            >
                <strong>{{ user.fullName }}</strong> ({{ user.username }})
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { UserInfo } from '@ethanrous/weblens-api'
import WeblensInput from '../atom/WeblensInput.vue'
import { onClickOutside, useFocusWithin } from '@vueuse/core'
import { useWeblensAPI } from '~/api/AllApi'

const search = ref<string>('')
const open = ref<boolean>(false)

const inputContainerRef = ref<HTMLDivElement>()
const focused = useFocusWithin(inputContainerRef)

onClickOutside(inputContainerRef, () => {
    open.value = false
})

watch(focused.focused, (newVal) => {
    if (newVal) {
        open.value = true
    }
})

const props = defineProps<{
    excludeFn?: (user: UserInfo) => boolean
}>()

const { data: users } = useAsyncData(
    'userSearch-' + search.value,
    async () => {
        if (search.value.length < 2) {
            return []
        }

        return useWeblensAPI()
            .UsersAPI.searchUsers(search.value)
            .then((response) => {
                const users = response.data
                if (props.excludeFn) {
                    return users.filter(props.excludeFn)
                }
                return users
            })
            .catch((error) => {
                console.error('Error fetching user search results:', error)
                return []
            })
    },
    { watch: [search], getCachedData: () => undefined },
)

const emit = defineEmits<{
    (e: 'select:user', value: UserInfo): void
}>()
</script>
