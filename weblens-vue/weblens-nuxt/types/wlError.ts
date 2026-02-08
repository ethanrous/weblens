import type { NuxtError } from '#app'
import type { AxiosError } from 'axios'

export class WLError {
    message?: string
    status?: number
    stack?: string

    constructor(error: Partial<WLError> | NuxtError | AxiosError) {
        if ('response' in error && error.response) {
            this.status = error.response.status
            this.message = error.response.data?.message || error.message
        } else if ('statusCode' in error && error.statusCode) {
            this.message = error.message
            this.status = error.statusCode
        }

        this.stack = error.stack
    }
}
