// Package agno provides CGO bindings for the AGNO image processing library.
package agno

/*
#cgo LDFLAGS: -L${SRCDIR}/lib -lagno -lstdc++
#cgo CFLAGS: -I${SRCDIR}/lib
#include "lib/agno.h"
*/
import "C"
import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
)

// Image represents an image loaded via the agno library.
type Image struct {
	img   *C.struct_AgnoImage
	mu    sync.Mutex
	freed bool
}

// ExifData holds EXIF metadata from an image.
type ExifData struct {
	data uint
	len  C.size_t
	typ  C.int16_t
}

func init() {
	C.init_agno()
}

// Dimensions returns the width and height of the image.
func (img *Image) Dimensions() (width, height int) {
	return int(img.img.width), int(img.img.height)
}

// func (img *Image) Resize(newWidth, newHeight int) error {
// 	newImg := C.resize_image(img.img, C.size_t(newWidth), C.size_t(newHeight))
//
// 	img.img = newImg
//
// 	return nil
// }

// Resize scales the image by the given factor.
func (img *Image) Resize(scale float64) error {
	newWidth := int(float64(img.img.width) * scale)
	newHeight := int(float64(img.img.height) * scale)

	newImg := C.resize_image(img.img, C.size_t(newWidth), C.size_t(newHeight)) //nolint:nlreturn
	img.img = newImg

	return nil
}

// GetExifValue retrieves an EXIF value of type T from the image.
func GetExifValue[T any](img *Image, exifTag int) (T, error) {
	v := img.getExifValue(exifTag)
	if v == -1 {
		var zero T

		return zero, errors.Errorf("failed to get exif value, unknown type")
	}

	if val, ok := v.(T); ok {
		return val, nil
	}

	var zero T

	return zero, errors.Errorf("failed to convert exif value to T (%T): value is %T with value %s", zero, v, v)
}

// ImageByFilepath loads an image from the given file path.
func ImageByFilepath(path string) (*Image, error) {
	cPathStr := C.CString(path)

	defer C.free(unsafe.Pointer(cPathStr)) //nolint:nlreturn

	cAgnoImg := C.load_image_from_path(cPathStr, C.size_t(len(path)))

	if cAgnoImg == nil || cAgnoImg.len == 0 || cAgnoImg.width == 0 || cAgnoImg.height == 0 {
		return nil, errors.Errorf("load_image_from_path returned nil loading [%s]", path)
	}

	img := &Image{img: cAgnoImg}
	runtime.SetFinalizer(img, func(img *Image) {
		img.Free()
	})

	return img, nil
}

// WriteWebp writes the image to a WebP file at the given path.
func WriteWebp(path string, img *Image) error {
	cPathStr := C.CString(path)

	defer C.free(unsafe.Pointer(cPathStr)) //nolint:nlreturn

	img.mu.Lock()
	defer img.mu.Unlock()

	C.write_agno_image_to_webp(cPathStr, C.size_t(len(path)), img.img)

	return nil
}

// GetImageSize returns the width and height of the image.
func GetImageSize(img *Image) (int, int) {
	img.mu.Lock()
	defer img.mu.Unlock()

	width := int(img.img.width)
	height := int(img.img.height)

	return width, height
}

// Free releases the image resources.
func (img *Image) Free() {
	img.mu.Lock()
	defer img.mu.Unlock()

	if img.freed {
		return
	}

	if img.img != nil {
		C.free_agno_image(img.img)
	}

	img.freed = true
}

func (img *Image) getExifValue(exifTag int) any {
	img.mu.Lock()
	defer img.mu.Unlock()
	v := C.get_exif_value(img.img, C.int16_t(exifTag)) //nolint:nlreturn

	switch v.typ {
	case 1: // BYTE
		log.GlobalLogger().Info().Msgf("EXIF type BYTE")
	case 2: // ASCII
		s := make([]byte, v.len)
		for i := 0; i < int(v.len); i++ {
			s[i] = *(*byte)(unsafe.Add(unsafe.Pointer(v.data), i))
		}

		return string(s)
	case 3: // SHORT
		return *(*uint32)(unsafe.Pointer(v.data))
	case 4: // LONG
		log.GlobalLogger().Info().Msgf("EXIF type LONG")
	case 5: // RATIONAL
		log.GlobalLogger().Info().Msgf("EXIF type RATIONAL")
	case 7: // UNDEFINED
		log.GlobalLogger().Info().Msgf("EXIF type UNDEFINED")
	case 9: // SLONG
		log.GlobalLogger().Info().Msgf("EXIF type SLONG")
	case 10: // SRATIONAL
		log.GlobalLogger().Info().Msgf("EXIF type SRATIONAL")
	default:
		log.GlobalLogger().Info().Msgf("EXIF type UNKNOWN %d", v.typ)
	}

	return -1
}
