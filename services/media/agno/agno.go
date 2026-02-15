// Package agno provides CGO bindings for the AGNO image processing library.
package agno

/*
#cgo LDFLAGS: -L${SRCDIR}/lib -lagno -lstdc++ -lm
#cgo darwin LDFLAGS: -framework Metal -framework QuartzCore -framework CoreGraphics
#cgo CFLAGS: -I${SRCDIR}/lib
#include "lib/agno.h"
*/
import "C"
import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
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

	start := time.Now()
	newImg := C.resize_image(img.img, C.size_t(newWidth), C.size_t(newHeight)) //nolint:nlreturn
	log.GlobalLogger().Debug().Msgf("Resized image to %dx%d in %s", newWidth, newHeight, time.Since(start))
	img.img = newImg

	return nil
}

// GetExifValue retrieves an EXIF value of type T from the image.
func GetExifValue[T any](img *Image, exifTag int) (T, error) {
	v := img.getExifValue(exifTag)

	if v == nil {
		var zero T

		return zero, wlerrors.Errorf("exif tag 0x%04x not found", exifTag)
	}

	if val, ok := v.(T); ok {
		return val, nil
	}

	var zero T

	return zero, wlerrors.Errorf("failed to convert exif value to %T: value is %T with value %v", zero, v, v)
}

// ImageByFilepath loads an image from the given file path.
func ImageByFilepath(path string) (*Image, error) {
	cPathStr := C.CString(path)

	defer C.free(unsafe.Pointer(cPathStr)) //nolint:nlreturn

	start := time.Now()
	cAgnoImg := C.load_image_from_path(cPathStr, C.size_t(len(path)))
	log.GlobalLogger().Debug().Msgf("Loaded image from path [%s] in %s", path, time.Since(start))

	if cAgnoImg == nil || cAgnoImg.len == 0 || cAgnoImg.width == 0 || cAgnoImg.height == 0 {
		return nil, wlerrors.Errorf("load_image_from_path returned nil loading [%s]", path)
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

// WriteJpeg encodes the image as JPEG with the given quality (1-100) and returns the bytes.
func WriteJpeg(img *Image, quality int) ([]byte, error) {
	img.mu.Lock()
	defer img.mu.Unlock()

	buf := C.write_agno_image_to_jpeg_buffer(img.img, C.uint8_t(quality)) //nolint:nlreturn
	if buf.data == nil {
		return nil, wlerrors.Errorf("failed to encode JPEG")
	}

	bs := C.GoBytes(unsafe.Pointer(buf.data), C.int(buf.len)) //nolint:nlreturn
	C.free_agno_buffer(buf)

	return bs, nil
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

// GetGPSCoordinates extracts GPS coordinates from the image's EXIF data.
// Returns [lat, lon] as decimal degrees, or an error if GPS data is not available.
func GetGPSCoordinates(img *Image) ([2]float64, error) {
	img.mu.Lock()
	defer img.mu.Unlock()

	gps := C.get_gps_coordinates(img.img) //nolint:nlreturn
	if gps.valid == 0 {
		return [2]float64{}, wlerrors.Errorf("no GPS coordinates in EXIF data")
	}

	return [2]float64{float64(gps.lat), float64(gps.lon)}, nil
}

func (img *Image) getExifValue(exifTag int) any {
	img.mu.Lock()
	defer img.mu.Unlock()
	v := C.get_exif_value(img.img, C.int16_t(exifTag)) //nolint:nlreturn

	if v.len == 0 && v.data == nil {
		return nil
	}

	switch v.typ {
	case 1: // BYTE
		return int(*(*byte)(unsafe.Pointer(v.data)))
	case 2: // ASCII
		s := make([]byte, v.len)
		for i := 0; i < int(v.len); i++ {
			s[i] = *(*byte)(unsafe.Add(unsafe.Pointer(v.data), i))
		}

		return string(s)
	case 3: // SHORT
		return int(*(*uint32)(unsafe.Pointer(v.data)))
	case 4: // LONG
		return int(*(*uint32)(unsafe.Pointer(v.data)))
	case 5: // RATIONAL
		num := *(*uint32)(unsafe.Pointer(v.data))
		den := *(*uint32)(unsafe.Add(unsafe.Pointer(v.data), 4))
		if den == 0 {
			return float64(0)
		}

		return float64(num) / float64(den)
	case 7: // UNDEFINED
		s := make([]byte, v.len)
		for i := 0; i < int(v.len); i++ {
			s[i] = *(*byte)(unsafe.Add(unsafe.Pointer(v.data), i))
		}

		return string(s)
	case 9: // SLONG
		return int(*(*int32)(unsafe.Pointer(v.data)))
	case 10: // SRATIONAL
		num := *(*int32)(unsafe.Pointer(v.data))
		den := *(*int32)(unsafe.Add(unsafe.Pointer(v.data), 4))
		if den == 0 {
			return float64(0)
		}

		return float64(num) / float64(den)
	default:
		log.GlobalLogger().Warn().Msgf("Unknown EXIF type %d for tag 0x%04x", v.typ, exifTag)

		return nil
	}
}
