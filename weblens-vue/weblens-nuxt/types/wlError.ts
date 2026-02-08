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
            this.message =
                data &&
                typeof data === 'object' &&
                'message' in data &&
                typeof (data as Record<string, unknown>).message === 'string'
                    ? ((data as Record<string, unknown>).message as string)
                    : error.message
        } else if ('statusCode' in error && error.statusCode) {
            this.message = error.message
            this.status = error.statusCode
        }

        this.stack = error.stack
    }
}
