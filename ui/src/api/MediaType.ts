// import { mediaType } from "../types/Types";
// import { fetchMediaTypes } from "./ApiFetch";

// var MediaTypes: Map<string, mediaType> = new Map<string, mediaType>();

// export async function VerifyMediaTypeMap() {
//     if (MediaTypes.size === 0) {
//         fetchMediaTypes().then((mt) => {
//             const mimes: string[] = Array.from(Object.keys(mt));
//             for (const mime of mimes) {
//                 MediaTypes.set(mime, mt[mime]);
//             }
//         });
//     }
// }

// export function GetMediaType(mimeType: string): mediaType {
//     return MediaTypes.get(mimeType);
// }
