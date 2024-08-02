export interface ShareInfo {
    id: string;
    accessors?: string[];
    expires?: string;
    public?: boolean;
    fileId?: string;
    shareName?: string;
    wormhole?: boolean;
}

export class WeblensShare {
    id: string;
    accessors: string[];
    expires: string;
    public: boolean;
    fileId: string;
    shareName: string;
    wormhole: boolean;

    constructor(init: ShareInfo) {
        if (!init) {
            console.error('Attempt to init share with no data');
            return;
        }
        Object.assign(this, init);
        if (!this.accessors) {
            this.accessors = [];
        }
    }

    Id(): string {
        return this.id;
    }

    IsPublic() {
        return this.public;
    }

    IsWormhole() {
        return this.wormhole;
    }

    GetFileId(): string {
        return this.fileId;
    }

    GetAccessors(): string[] {
        return this.accessors;
    }

    GetPublicLink(): string {
        return `${window.location.origin}/files/share/${this.id}/${this.fileId}`;
    }
}
