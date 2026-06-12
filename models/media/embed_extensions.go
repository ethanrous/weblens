package media

import "strings"

// EmbedEligible reports whether a file with this extension (with or without a
// leading dot, any case) should be processed by the semantic-embedding pipeline.
func EmbedEligible(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	return ParseExtension(ext).IsEmbeddable()
}
