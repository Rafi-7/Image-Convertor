package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <image-file-path1> <image-file-path2>...")
		return
	}

	var waitGroup sync.WaitGroup
	for _, imagePath := range os.Args[1:] {
		waitGroup.Add(1)
		go func(imagePath string) {
			defer waitGroup.Done()
			if err := ConvertToBW(imagePath); err != nil {
				fmt.Printf("Failed to process %s : %v\n", imagePath, err)
			}
		}(imagePath)
	}
	waitGroup.Wait()
}

func ConvertToBW(imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("Error decoding image file %s : %v\n", imagePath, err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("Error decoding image file : %v", err)
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	pixels := make(chan Pixel)
	done := make(chan bool)

	go func() {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				pixelColor := img.At(x, y)
				grayColor := color.GrayModel.Convert(pixelColor).(color.Gray)
				pixels <- Pixel{x, y, grayColor}
			}
		}
		close(pixels)
	}()

	go func() {
		for pixel := range pixels {
			grayImg.SetGray(pixel.x, pixel.y, pixel.color)
		}
		done <- true
	}()
	<-done

	outputPath := generationOutputPath(imagePath, format)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Error creating output file %s: %v\n", outputPath, err)
	}
	defer outputFile.Close()

	switch format {
	case "jpeg":
		if err := jpeg.Encode(outputFile, grayImg, nil); err != nil {
			return fmt.Errorf("Error Eencoding output file %s: %v\n", outputPath, err)
		}
	case "png":
		if err := png.Encode(outputFile, grayImg); err != nil {
			return fmt.Errorf("Error Eencoding output file %s: %v\n", outputPath, err)
		}
	default:
		return fmt.Errorf("Unsupported image format %s: %v\n", imagePath, format)
	}

	fmt.Printf("Image %s converted to BW %s\n", imagePath, outputPath)
	return nil
}

func generationOutputPath(inputPath, _ string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	return fmt.Sprintf("%s_bw%s", base, ext)
}

type Pixel struct {
	x, y  int
	color color.Gray
}
