package media

import "strings"

// embedEligibleExtensions are the file extensions the semantic-embedding
// pipeline will process. Documents, code, and plaintext are extracted as text;
// the listed raster images are OCR'd. Registry image types (jpg/png/heic) are
// included so the background extract task indexes any text within them; the
// scan path tells text from visual embedding via MType.SupportsImgRecog.
var embedEligibleExtensions = map[string]bool{
	"txt": true, "md": true, "csv": true, "log": true,
	"json": true, "yaml": true, "yml": true,
	"go": true, "py": true, "js": true, "ts": true, "tsx": true,
	"vue": true, "rs": true, "java": true, "c": true, "cpp": true,
	"h": true, "hpp": true, "sh": true, "rb": true, "kt": true, "swift": true,
	"pdf":  true,
	"docx": true, "xlsx": true, "pptx": true,
	"jpg": true, "jpeg": true, "png": true, "heic": true,
	"tif": true, "tiff": true, "bmp": true,
}

// EmbedEligible reports whether a file with this extension (with or without a
// leading dot, any case) should be processed by the semantic-embedding pipeline.
func EmbedEligible(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")

	return embedEligibleExtensions[ext]
}
