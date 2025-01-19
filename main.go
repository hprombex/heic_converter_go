/*
Copyright (c) 2025 hprombex

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
OR OTHER DEALINGS IN THE SOFTWARE.

Author: hprombex

HEIC Converter for converting .HEIC images to other formats like JPEG or PNG
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/strukturag/libheif/go/heif"
)

// Saves an image as a JPEG file with the specified quality and filename.
func saveJPEG(img image.Image, filename string, quality int) {
	var out bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&out, img, opts); err != nil {
		fmt.Printf("Could not encode image as JPEG: %s\n", err)
	} else {
		if err := os.WriteFile(filename, out.Bytes(), 0644); err != nil {
			fmt.Printf("Could not save JPEG image as %s: %s\n", filename, err)
		} else {
			fmt.Printf("HEIC image saved as %s\n", filename)
		}
	}
}

// Saves an image as a PNG file with the specified filename.
func savePNG(img image.Image, filename string) {
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		fmt.Printf("Could not encode image as PNG: %s\n", err)
	} else {
		if err := os.WriteFile(filename, out.Bytes(), 0644); err != nil {
			fmt.Printf("Could not save PNG image as %s: %s\n", filename, err)
		} else {
			fmt.Printf("HEIC image saved as %s\n", filename)
		}
	}
}

// Converts a HEIC file to JPEG or PNG and optionally deletes the original.
func convertHeic(file string, outputPath string, format string, quality int, deleteOriginal bool, wg *sync.WaitGroup, start <-chan struct{}, done chan struct{}) {
	defer wg.Done()
	<-start // Wait for the start signal for all workers

	c, err := heif.NewContext()
	if err != nil {
		fmt.Printf("Could not create context: %s\n", err)
		return
	}

	if err := c.ReadFromFile(file); err != nil {
		fmt.Printf("Could not read file %s: %s\n", file, err)
		return
	}

	handle, err := c.GetPrimaryImageHandle()
	if err != nil {
		fmt.Printf("Could not get primary image: %s\n", err)
		return
	}

	fmt.Printf("Converting file: %s image size: %v × %v\n", file, handle.GetWidth(), handle.GetHeight())
	img, err := handle.DecodeImage(heif.ColorspaceUndefined, heif.ChromaUndefined, nil)
	if err != nil {
		fmt.Printf("Could not decode image: %s\n", err)
	} else if i, err := img.GetImage(); err != nil {
		fmt.Printf("Could not get image: %s\n", err)
	} else {
		outFilename := strings.Replace(file, ".", "_", 1)
		if outputPath != "" {
			filename := filepath.Base(file)
			outFilename = outputPath + strings.Replace(filename, ".", "_", 1)
		}

		switch format {
		case "jpeg":
			saveJPEG(i, outFilename + ".jpg", quality)
		case "png":
			savePNG(i, outFilename + ".png")
		default:
			fmt.Printf("Unsupported format: %s\n", format)
			return
		}
	}

	if deleteOriginal {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Failed to delete original file %s: %v", file, err)
		} else {
			fmt.Printf("Deleted original file: %s", file)
		}
	}

	<-done // Release a slot in the semaphore
}

// Finds all HEIC files in a directory and returns their paths.
func FindHeicFiles(directory string) ([]string, error) {
	var heicFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(strings.ToLower(info.Name())) == ".heic" {
			heicFiles = append(heicFiles, path)
		}
		return nil
	})
	return heicFiles, err
}

func main() {
	inputFile := flag.String("input_file", "", "Path to a single .HEIC file to be converted.")
	inputDir := flag.String("input_dir", "", "Path to a directory containing .HEIC files.")
	outputPath := flag.String("output_path", "", "Path to the output file or directory.")
	deleteOriginal := flag.Bool("delete", false, "Delete the original file after conversion.")
	format := flag.String("format", "jpeg", "Output image format (jpeg or png).")
	quality := flag.Int("quality", 80, "Quality of the output image (1-100).")

	flag.Parse()

	NumCPUs := runtime.NumCPU() // Maximum number of goroutines running concurrently
	fmt.Printf("Number of CPUs: %d\n", NumCPUs)

	done := make(chan struct{}, NumCPUs)
	start := make(chan struct{})
	var wg sync.WaitGroup

	if *inputFile != "" {
		if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
			fmt.Printf("Input file '%s' does not exist.", *inputFile)
		}
		wg.Add(1)
		done <- struct{}{}
		go convertHeic(*inputFile, *outputPath, *format, *quality, *deleteOriginal, &wg, start, done)
	} else if *inputDir != "" {
		if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
			fmt.Printf("Input directory '%s' does not exist.", *inputDir)
		}
		files, err := FindHeicFiles(*inputDir)
		if err != nil {
			fmt.Printf("Error finding HEIC files: %v", err)
		}
		for _, file := range files {
			wg.Add(1)
			go func() {
				done <- struct{}{} //Reserve a slot in the semaphore
				convertHeic(file, *outputPath, *format, *quality, *deleteOriginal, &wg, start, done)
			}()
		}
	} else {
		fmt.Println("Either --input_file or --input_dir must be specified.")
	}

	close(start) // Send the start signal to all workers
	wg.Wait()    // Wait for all workers to finish

	fmt.Println("All conversions completed.")
}
