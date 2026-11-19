package imageproc

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func jpegOf(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

func size(t *testing.T, data []byte) (int, int) {
	t.Helper()
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	require.NoError(t, err)
	return cfg.Width, cfg.Height
}

// withOrientation inserts an EXIF segment with the given orientation after the SOI marker.
func withOrientation(data []byte, orientation uint16) []byte {
	tiff := []byte("MM\x00\x2a\x00\x00\x00\x08")
	tiff = binary.BigEndian.AppendUint16(tiff, 1) // one IFD entry
	tiff = binary.BigEndian.AppendUint16(tiff, 0x0112)
	tiff = binary.BigEndian.AppendUint16(tiff, 3) // SHORT
	tiff = binary.BigEndian.AppendUint32(tiff, 1)
	tiff = binary.BigEndian.AppendUint16(tiff, orientation)
	tiff = append(tiff, 0, 0, 0, 0, 0, 0)
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1}
	seg = binary.BigEndian.AppendUint16(seg, uint16(len(payload)+2))
	seg = append(seg, payload...)
	out := append([]byte{}, data[:2]...)
	out = append(out, seg...)
	return append(out, data[2:]...)
}

func TestCopies_LargePhoto_ScaledToWidths(t *testing.T) {
	small, large, err := Copies(jpegOf(t, image.NewRGBA(image.Rect(0, 0, 4000, 2000))))
	require.NoError(t, err)
	w, h := size(t, large)
	assert.Equal(t, [2]int{1600, 800}, [2]int{w, h})
	w, h = size(t, small)
	assert.Equal(t, [2]int{800, 400}, [2]int{w, h})
}

func TestCopies_NarrowPhoto_NotEnlarged(t *testing.T) {
	small, large, err := Copies(jpegOf(t, image.NewRGBA(image.Rect(0, 0, 600, 900))))
	require.NoError(t, err)
	w, _ := size(t, large)
	assert.Equal(t, 600, w)
	w, _ = size(t, small)
	assert.Equal(t, 600, w)
}

func TestCopies_RotatedPhoto_TurnedUpright(t *testing.T) {
	// Stored landscape 3000×2000 with orientation 6: upright it is portrait 2000×3000.
	img := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	// A white top-left corner; turned 90° clockwise it ends up top-right.
	for y := 0; y < 400; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.White)
		}
	}
	_, large, err := Copies(withOrientation(jpegOf(t, img), 6))
	require.NoError(t, err)
	w, h := size(t, large)
	assert.Equal(t, [2]int{1600, 2400}, [2]int{w, h})

	out, err := jpeg.Decode(bytes.NewReader(large))
	require.NoError(t, err)
	r, _, _, _ := out.At(w-50, 50).RGBA()
	assert.Greater(t, r>>8, uint32(200), "top-right corner should be white")
	r, _, _, _ = out.At(50, 50).RGBA()
	assert.Less(t, r>>8, uint32(50), "top-left corner should be dark")
}

func TestCopies_NotAnImage_ReturnsError(t *testing.T) {
	_, _, err := Copies([]byte("<svg xmlns='http://www.w3.org/2000/svg'/>"))
	assert.Error(t, err)
}

func TestExifOrientation_Absent_IsUpright(t *testing.T) {
	assert.Equal(t, 1, exifOrientation(jpegOf(t, image.NewRGBA(image.Rect(0, 0, 10, 10)))))
	assert.Equal(t, 8, exifOrientation(withOrientation(jpegOf(t, image.NewRGBA(image.Rect(0, 0, 10, 10))), 8)))
}
