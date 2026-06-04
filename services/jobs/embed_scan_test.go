package jobs

import "testing"

func TestShouldExtractTextOnScan(t *testing.T) {
	cases := map[string]bool{
		".pdf":  true,
		"pdf":   true,
		".PDF":  true,
		".docx": true,
		".txt":  true,
		".md":   true,
		".jpg":  false,
		".jpeg": false,
		".png":  false,
		".heic": false,
		".mp4":  false, // not eligible at all
		"":      false,
	}
	for ext, want := range cases {
		if got := shouldExtractTextOnScan(ext); got != want {
			t.Errorf("shouldExtractTextOnScan(%q) = %v, want %v", ext, got, want)
		}
	}
}
