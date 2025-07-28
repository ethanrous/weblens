# MediaInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**contentId** | **string** | Hash of the file content, to ensure that the same files don\&#39;t get duplicated | [optional] [default to undefined]
**createDate** | **number** |  | [optional] [default to undefined]
**duration** | **number** | Total time, in milliseconds, of a video | [optional] [default to undefined]
**enabled** | **boolean** | If the media disabled. This can happen when the backing file(s) are deleted, but the media stays behind because it can be re-used if needed. | [optional] [default to undefined]
**fileIds** | **Array&lt;string&gt;** | Slices of files whos content hash to the contentId | [optional] [default to undefined]
**height** | **number** |  | [optional] [default to undefined]
**hidden** | **boolean** | If the media is hidden from the timeline TODO - make this per user | [optional] [default to undefined]
**imported** | **boolean** |  | [optional] [default to undefined]
**likedBy** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**location** | **Array&lt;number&gt;** |  | [optional] [default to undefined]
**mimeType** | **string** | Mime-type key of the media | [optional] [default to undefined]
**owner** | **string** | User who owns the file that resulted in this media being created | [optional] [default to undefined]
**pageCount** | **number** | Number of pages (typically 1, 0 in not a valid page count) | [optional] [default to undefined]
**recognitionTags** | **Array&lt;string&gt;** | Tags from the ML image scan so searching for particular objects in the images can be done | [optional] [default to undefined]
**width** | **number** | Full-res image dimensions | [optional] [default to undefined]

## Example

```typescript
import { MediaInfo } from './api';

const instance: MediaInfo = {
    contentId,
    createDate,
    duration,
    enabled,
    fileIds,
    height,
    hidden,
    imported,
    likedBy,
    location,
    mimeType,
    owner,
    pageCount,
    recognitionTags,
    width,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
