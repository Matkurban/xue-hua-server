package rustfs

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	formatPng  = "png"
	formatJpeg = "jpeg"
	formatJpg  = "jpg"
	formatGif  = "gif"
)

func ImageStat(reader io.Reader) (image.Image, string, error) {
	return image.Decode(reader)
}

func ImageWidthHeight(img image.Image) (int, int) {
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()
	if imgWidth <= 0 || imgHeight <= 0 || (maxWidth <= 0 && maxHeight <= 0) {
		return img
	}

	var scale float64
	switch {
	case maxWidth > 0 && maxHeight > 0:
		scaleW := float64(maxWidth) / float64(imgWidth)
		scaleH := float64(maxHeight) / float64(imgHeight)
		scale = min(scaleW, scaleH)
	case maxWidth > 0:
		scale = float64(maxWidth) / float64(imgWidth)
	default:
		scale = float64(maxHeight) / float64(imgHeight)
	}

	thumbnailWidth := max(int(float64(imgWidth)*scale), 1)
	thumbnailHeight := max(int(float64(imgHeight)*scale), 1)

	thumbnail := image.NewRGBA(image.Rect(0, 0, thumbnailWidth, thumbnailHeight))
	for y := 0; y < thumbnailHeight; y++ {
		srcY := bounds.Min.Y + int(float64(y)/scale)
		for x := 0; x < thumbnailWidth; x++ {
			srcX := bounds.Min.X + int(float64(x)/scale)
			thumbnail.Set(x, y, img.At(srcX, srcY))
		}
	}
	return thumbnail
}

func getThumbnailSize(img image.Image) (int, int) {
	w, h := ImageWidthHeight(img)
	return min(w, 640), min(h, 640)
}
