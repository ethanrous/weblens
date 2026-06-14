package media

import "strings"

// EmbedEligible reports whether a file with this extension (with or without a
// leading dot, any case) should be processed by the semantic-embedding pipeline.
func EmbedEligible(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	return ParseExtension(ext).IsEmbeddable()
}

// TextEmbedEligible reports whether a file with this extension should get text
// extraction (OCR/parsing) embeddings. Photo types are excluded: they are embedded
// visually during scan, and OCR on photos yields junk like watermark text.
func TextEmbedEligible(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	return ParseExtension(ext).IsEmbeddable() && !ParseExtension(ext).SupportsImgRecog()
}
