package agno

/*
#cgo LDFLAGS: -L/agno/lib -lagno -lstdc++
#include "/agno/lib/agno.h"
*/
import "C"
import (
	"runtime"
	"sync"
	"unsafe"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
)

type Image struct {
	img   *C.struct_AgnoImage
	mu    sync.Mutex
	freed bool
}

type ExifData struct {
	data uint
	len  C.size_t
	typ  C.int16_t
}

func init() {
	C.init_agno()
}

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

func (img *Image) Resize(scale float64) error {
	newWidth := int(float64(img.img.width) * scale)
	newHeight := int(float64(img.img.height) * scale)

	newImg := C.resize_image(img.img, C.size_t(newWidth), C.size_t(newHeight))

	img.img = newImg

	return nil
}

func (img *Image) getExifValue(exifTag int) any {
	img.mu.Lock()
	defer img.mu.Unlock()

	v := C.get_exif_value(img.img, C.int16_t(exifTag))

	switch v.typ {
	case 1: // BYTE
		log.GlobalLogger().Info().Msgf("EXIF type BYTE")
	case 2: // ASCII
		s := make([]byte, v.len)
		for i := 0; i < int(v.len); i++ {
			s[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(v.data)) + uintptr(i)))
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
	c_path_str := C.CString(path)

	defer C.free(unsafe.Pointer(c_path_str))

	c_agno_img := C.load_image_from_path(c_path_str, C.size_t(len(path)))

	if c_agno_img == nil || c_agno_img.len == 0 || c_agno_img.width == 0 || c_agno_img.height == 0 {
		return nil, errors.Errorf("load_image_from_path returned nil loading [%s]", path)
	}

	img := &Image{img: c_agno_img}
	runtime.SetFinalizer(img, func(img *Image) {
		img.Free()
	})

	return img, nil
}

func WriteWebp(path string, img *Image) error {
	c_path_str := C.CString(path)

	defer C.free(unsafe.Pointer(c_path_str))

	img.mu.Lock()
	defer img.mu.Unlock()

	C.write_agno_image_to_webp(c_path_str, C.size_t(len(path)), img.img)

	return nil
}

func GetImageSize(img *Image) (int, int) {
	img.mu.Lock()
	defer img.mu.Unlock()

	width := int(img.img.width)
	height := int(img.img.height)

	return width, height
}

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
