import type { NuxtError } from '#app'
import type { AxiosError } from 'axios'

export class WLError {
    message?: string
    status?: number
    stack?: string

    constructor(error: Partial<WLError> | NuxtError | AxiosError) {
        if ('response' in error && error.response) {
            this.status = error.response.status

            const data = error.response.data
            if (data && typeof data === 'object') {
                this.message =
                    (data as Record<string, string>).message ?? (data as Record<string, string>).error ?? error.message
            } else {
                this.message = error.message
            }
        } else if ('status' in error && error.status) {
            this.message = error.message
            this.status = error.status
        }

        this.stack = error.stack
    }
}
