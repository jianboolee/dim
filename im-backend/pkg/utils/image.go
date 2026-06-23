package utils

import (
	"errors"
	"image"
	_ "image/gif"  // 支持 GIF 图片
	_ "image/jpeg" // 支持 JPEG 图片
	_ "image/png"  // 支持 PNG 图片
	"mime/multipart"
	"os"
)

// GetImageDimensions 从上传文件中获取图片的宽高
func GetImageDimensions(file multipart.File) (width, height int, err error) {
	// 解码图片文件，获取宽高
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, errors.New("failed to decode image: " + err.Error())
	}

	// 获取图片的宽度和高度
	width = img.Bounds().Dx()
	height = img.Bounds().Dy()

	return width, height, nil
}

// GetImageDimensionsFromPath 从文件路径中获取图片的宽高
func GetImageDimensionsFromPath(filePath string) (width, height int, err error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, errors.New("failed to open file: " + err.Error())
	}
	defer file.Close()

	// 解码图片文件
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, 0, errors.New("failed to decode image: " + err.Error())
	}

	// 获取图片的宽度和高度
	width = img.Bounds().Dx()
	height = img.Bounds().Dy()

	return width, height, nil
}
