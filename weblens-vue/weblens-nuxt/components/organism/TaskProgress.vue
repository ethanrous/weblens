<template>
    <div
        ref="tasksContainer"
        :class="{ 'my-2 flex h-max w-full flex-col': true }"
    >
        <div
            v-for="task of tasksArray"
            :key="task.taskID"
            :class="{
                'bg-card-background-primary group z-20 mb-3 h-max w-full rounded border p-2 transition-[width] last:mb-0': true,
            }"
        >
            <div :class="{ 'mb-1 flex flex-col justify-center': true }">
                <div :class="{ 'border-text-tertiary mb-1 flex w-full items-center border-b pb-1': true }">
                    <FileIcon
                        :class="{ 'mr-auto min-w-0 gap-0.5': true }"
                        :file="task.targetFile"
                        with-name
                    />

                    <span :class="{ 'text-text-secondary mr-auto text-nowrap': true }">Importing Media</span>

                    <IconX
                        size="20"
                        :class="{
                            'text-text-secondary hover:text-text-primary cursor-pointer opacity-0 transition group-hover:opacity-100': true,
                        }"
                        @click="removeTask(task.taskID)"
                    />
                </div>

                <div v-if="containerSize.width.value > 100">
                    <div
                        v-if="task.status !== TaskStatus.Completed"
                        :class="{ 'flex items-center py-1': true }"
                    >
                        <span
                            v-if="task.status === TaskStatus.InProgress"
                            :class="{ 'text-text-secondary border-text-tertiary text-nowrap': true }"
                        >
                            {{ task.countComplete }} / {{ task.countTotal }}
                        </span>

                        <span
                            v-if="task.status === TaskStatus.Failed"
                            :class="{
                                'text-text-secondary border-text-tertiary inline-flex w-full text-nowrap': true,
                            }"
                        >
                            {{ task.failCount }} of {{ task.countTotal }} files failed

                            <IconExclamationCircle
                                :class="{ 'text-danger ml-auto': true }"
                                size="20"
                            />
                        </span>

                        <span
                            v-else-if="task.status === TaskStatus.Canceled"
                            :class="{ 'text-text-secondary border-text-tertiary text-nowrap': true }"
                        >
                            Canceled
                        </span>

                        <span
                            v-if="task.status === TaskStatus.Pending"
                            :class="{
                                'text-text-secondary border-text-tertiary inline-flex w-full text-nowrap': true,
                            }"
                        >
                            Waiting to start

                            <Loader
                                size="12"
                                :class="{ 'mr-1 ml-auto': true }"
                            />
                        </span>
                    </div>

                    <span
                        v-if="task.status === TaskStatus.Completed && task.isScanDirectoryTask()"
                        :class="{ 'text-text-secondary my-1 text-nowrap': true }"
                    >
                        Imported in {{ humanDuration(task.executionTime() / (1000 * 1000)) }}
                    </span>
                </div>
            </div>

            <div :class="{ 'flex flex-row items-center': true }">
                <ProgressSquare
                    :class="{
                        'bg-background-primary h-2 w-full min-w-0 !shrink-1': true,
                    }"
                    :progress="task.percentComplete"
                    :failed="task.status === TaskStatus.Failed || task.status === TaskStatus.Canceled"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import ProgressSquare from '../atom/ProgressSquare.vue'
import FileIcon from '../atom/FileIcon.vue'
import { humanDuration } from '~/util/humanBytes'
import { IconX, IconExclamationCircle } from '@tabler/icons-vue'
import { useElementSize } from '@vueuse/core'
import { CancelTask } from '~/api/FileBrowserApi'
import { TaskStatus } from '~/types/task'
import Loader from '../atom/Loader.vue'

const tasksStore = useTasksStore()

const tasksContainer = ref<HTMLDivElement>()
const containerSize = useElementSize(tasksContainer)

const tasksArray = computed(() => {
    let tasksIttr = tasksStore.tasks?.values()
    if (!tasksIttr) {
        return []
    }

    tasksIttr = tasksIttr.filter((t) => {
        return t.isScanDirectoryTask()
    })

    const tasks = Array.from(tasksIttr)
    return tasks
})

function removeTask(taskID: string) {
    const task = tasksStore.tasks?.get(taskID)
    if (task?.status !== TaskStatus.InProgress) {
        tasksStore.removeTask(taskID)
        return
    } else {
        CancelTask(taskID)
    }
}
</script>
