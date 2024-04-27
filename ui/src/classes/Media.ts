import { useContext } from "react";
import { hideMedia } from "../api/ApiFetch";
import { AuthHeaderT, mediaType } from "../types/Types";
import { MediaTypeContext } from "../Context";

export interface MediaDataT {
    mediaId: string;
    mimeType?: string;

    fileIds?: string[];
    thumbnailCacheId?: string;
    fullresCacheIds?: string;
    blurHash?: string;
    owner?: string;
    mediaWidth?: number;
    mediaHeight?: number;
    createDate?: string;
    recognitionTags?: string[];
    pageCount?: number;
    hidden?: boolean;

    // Non-api props
    thumbnail?: ArrayBuffer;
    fullres?: ArrayBuffer;

    Previous?: WeblensMedia;
    Next?: WeblensMedia;
    selected?: boolean;
    mediaType?: mediaType;
    // Display: boolean
    ImgRef?: any;
}

class WeblensMedia {
    private data: MediaDataT;

    constructor(init: MediaDataT) {
        this.data = init;
    }

    Id(): string {
        return this.data.mediaId;
    }

    IsHidden(): boolean {
        return this.data.hidden;
    }

    HighestQualityLoaded(): "fullres" | "thumbnail" | "" {
        if (Boolean(this.data.fullres)) {
            return "fullres";
        } else if (Boolean(this.data.thumbnail)) {
            return "thumbnail";
        } else {
            return "";
        }
    }

    GetMediaType(): mediaType {
        const typeMap = useContext(MediaTypeContext);
        if (!this.data.mediaType) {
            this.data.mediaType = typeMap.get(this.data.mimeType);
        }
        return this.data.mediaType;
    }

    SetThumbnailBytes(buf: ArrayBuffer) {
        this.data.thumbnail = buf;
    }

    SetFullresBytes(buf: ArrayBuffer) {
        this.data.fullres = buf;
    }

    GetFileIds(): string[] {
        if (!this.data.fileIds) {
            return [];
        }

        return this.data.fileIds;
    }

    SetSelected(s: boolean) {
        this.data.selected = s;
    }

    IsSelected(): boolean {
        return this.data.selected;
    }

    IsDisplayable(): boolean {
        return this.GetMediaType().IsDisplayable;
    }

    SetImgRef(r) {
        this.data.ImgRef = r;
    }

    GetImgRef() {
        return this.data.ImgRef;
    }

    GetHeight(): number {
        return this.data.mediaHeight;
    }

    GetWidth(): number {
        return this.data.mediaWidth;
    }

    SetNextLink(next: WeblensMedia) {
        this.data.Next = next;
    }

    Next(): WeblensMedia {
        return this.data.Next;
    }

    SetPrevLink(prev: WeblensMedia) {
        this.data.Previous = prev;
    }

    Prev(): WeblensMedia {
        return this.data.Previous;
    }

    MatchRecogTag(searchTag: string): boolean {
        if (!this.data.recognitionTags) {
            return false;
        }

        return this.data.recognitionTags.includes(searchTag);
    }

    GetPageCount(): number {
        return this.data.pageCount;
    }

    async Hide(authHeader: AuthHeaderT) {
        this.data.hidden = true;
        return await hideMedia(this.data.mediaId, authHeader);
    }

    GetImgUrl(quality: "thumbnail" | "fullres"): string {
        if (quality == "thumbnail") {
            return URL.createObjectURL(new Blob([this.data.thumbnail]));
        } else if (quality == "fullres") {
            return URL.createObjectURL(new Blob([this.data.fullres]));
        }
    }
}

export default WeblensMedia;
