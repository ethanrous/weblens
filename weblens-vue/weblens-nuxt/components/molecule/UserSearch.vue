<template>
    <div
        ref="inputContainerRef"
        :class="{ 'relative w-full': true }"
    >
        <WeblensInput
            v-model:value="search"
            placeholder="Search Users..."
            :class="{ 'border-b-none rounded-b-none': users?.length && focused.focused.value }"
            @clear="search = ''"
        />
        <div
            v-if="users?.length"
            :class="{
                'bg-card-background-primary border-t-none absolute z-10 w-full rounded-b border pb-1 shadow': true,
            }"
        >
            <div
                v-for="user in users"
                :key="user.username"
                :class="{ 'hover:bg-card-background-hover m-1 cursor-pointer rounded p-2': true }"
                @click.stop.prevent="
                    () => {
                        emit('select:user', user)
                        search = ''
                    }
                "
            >
                <div>
                    <strong>{{ user.fullName }}</strong> ({{ user.username }})
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { UserInfo } from '@ethanrous/weblens-api'
import WeblensInput from '../atom/WeblensInput.vue'
import { useFocusWithin } from '@vueuse/core'
import { useWeblensAPI } from '~/api/AllApi'

const search = ref<string>('')
const inputContainerRef = ref<HTMLDivElement>()
const focused = useFocusWithin(inputContainerRef)

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
