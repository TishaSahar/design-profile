// Package imageproc makes the web copies of project photos: the site shows
// these instead of the uploaded originals, which are often several megabytes.
package imageproc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // decoders for image.Decode
	"image/jpeg"
	_ "image/png"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// Widths of the copies, in pixels. Narrower photos are not enlarged.
const (
	SmallWidth = 800  // grid of a project, cards, admin thumbnails
	LargeWidth = 1600 // the same places on large or high-density screens
	quality    = 82
)

// Copies decodes a photo and returns JPEG copies at most SmallWidth and
// LargeWidth pixels wide, turned upright by the EXIF orientation of a JPEG.
// Transparent areas become white. GIF and SVG have no copies: an animation
// would be lost, and SVG is not a raster image.
func Copies(data []byte) (small, large []byte, err error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("decode photo: %w", err)
	}
	if format == "gif" {
		return nil, nil, fmt.Errorf("gif photos are served as uploaded")
	}
	orientation := 1
	if format == "jpeg" {
		orientation = exifOrientation(data)
	}
	// Orientations 5–8 turn the photo by 90°: its upright width is the stored height.
	sideways := orientation >= 5 && orientation <= 8

	largeImg := orient(fit(img, LargeWidth, sideways), orientation)
	if large, err = encode(largeImg); err != nil {
		return nil, nil, err
	}
	if small, err = encode(fit(largeImg, SmallWidth, false)); err != nil {
		return nil, nil, err
	}
	return small, large, nil
}

// fit returns src scaled down so that its width (its height when sideways) is at
// most maxWidth, drawn over white.
func fit(src image.Image, maxWidth int, sideways bool) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	limited := w
	if sideways {
		limited = h
	}
	if limited > maxWidth {
		w = max(1, (w*maxWidth+limited/2)/limited)
		h = max(1, (h*maxWidth+limited/2)/limited)
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	if w == b.Dx() && h == b.Dy() {
		draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	} else {
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	}
	return dst
}

func encode(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode photo: %w", err)
	}
	return buf.Bytes(), nil
}

// orient turns an image upright according to an EXIF orientation (1–8).
func orient(src *image.RGBA, orientation int) *image.RGBA {
	if orientation < 2 || orientation > 8 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var dx, dy int
			switch orientation {
			case 2: // mirror left-right
				dx, dy = w-1-x, y
			case 3: // rotate 180°
				dx, dy = w-1-x, h-1-y
			case 4: // mirror top-bottom
				dx, dy = x, h-1-y
			case 5: // transpose
				dx, dy = y, x
			case 6: // rotate 90° clockwise
				dx, dy = h-1-y, x
			case 7: // transverse
				dx, dy = h-1-y, w-1-x
			case 8: // rotate 90° counter-clockwise
				dx, dy = y, w-1-x
			}
			dst.SetRGBA(dx, dy, src.RGBAAt(x, y))
		}
	}
	return dst
}

// exifOrientation reads the orientation tag of a JPEG; 1 (upright) when absent.
func exifOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}
	for i := 2; i+4 <= len(data); {
		if data[i] != 0xFF {
			return 1
		}
		marker := data[i+1]
		if marker == 0xDA || marker == 0xD9 { // image data starts: no EXIF before it
			return 1
		}
		size := int(binary.BigEndian.Uint16(data[i+2:]))
		if size < 2 || i+2+size > len(data) {
			return 1
		}
		segment := data[i+4 : i+2+size]
		if marker == 0xE1 && len(segment) >= 6 && string(segment[:6]) == "Exif\x00\x00" {
			return tiffOrientation(segment[6:])
		}
		i += 2 + size
	}
	return 1
}

func tiffOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	ifd := int(order.Uint32(tiff[4:]))
	if ifd+2 > len(tiff) {
		return 1
	}
	count := int(order.Uint16(tiff[ifd:]))
	for n := 0; n < count; n++ {
		entry := ifd + 2 + n*12
		if entry+12 > len(tiff) {
			return 1
		}
		if order.Uint16(tiff[entry:]) == 0x0112 { // Orientation, a SHORT
			if v := int(order.Uint16(tiff[entry+8:])); v >= 1 && v <= 8 {
				return v
			}
			return 1
		}
	}
	return 1
}
