<template>
    <teleport to="body">
        <div
            v-if="confirmDialog.isOpen"
            :class="{
                'fixed inset-0 z-99 flex items-center justify-center bg-black/50': true,
            }"
        >
            <div
                :class="{
                    'bg-card-background-primary w-11/12 max-w-md rounded border p-5 shadow-lg': true,
                }"
            >
                <h2 :class="{ 'mb-4 text-xl font-semibold': true }">Are you sure?</h2>
                <p :class="{ 'text-text-secondary mb-6 text-sm': true }">
                    Are you sure you want to {{ confirmDialog.verb }} {{ confirmDialog.itemName }}?
                </p>
                <div :class="{ 'flex justify-end gap-4': true }">
                    <WeblensButton
                        label="Cancel"
                        flavor="secondary"
                        @click="confirmDialog.cancel()"
                    />
                    <WeblensButton
                        label="Confirm"
                        flavor="danger"
                        @click="confirmDialog.accept()"
                    />
                </div>
            </div>
        </div>
    </teleport>
</template>

<script setup lang="ts">
import useConfirmDialog from '~/stores/confirm'
import WeblensButton from '../atom/WeblensButton.vue'
const confirmDialog = useConfirmDialog()
</script>
