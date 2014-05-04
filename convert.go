package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"image/png"
)

type Sci struct {
	Left   int
	Top    int
	Bottom int
	Right  int
	Source string
}

func (self Sci) String() string {
	return "border.left: " + strconv.Itoa(self.Left) + "\nborder.right: " + strconv.Itoa(self.Right) + "\nborder.bottom: " + strconv.Itoa(self.Bottom) + "\nborder.top: " + strconv.Itoa(self.Top) + "\nsource: " + self.Source
}

func main() {
	var imgFlag string
	var verbose bool
	flag.StringVar(&imgFlag, "img", "", "The 9patch image")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()

	imgDir := filepath.Dir(imgFlag)
	imgBase := filepath.Base(imgFlag)

	if !strings.HasSuffix(imgBase, ".9.png") {
		log.Fatal("Provided image doesn't have .9.png extention")
	}

	imgNoExtBase := imgBase[0 : len(imgBase)-6]
	if verbose {
		fmt.Print("Image Path: " + imgDir + " Base " + imgBase + "\n")
	}

	fi, err := os.Open(imgFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		log.Fatal(err)
	}
	fi.Close()

	bounds := img.Bounds()
	sci := Sci{-1, -1, -1, -1, imgNoExtBase + ".png"}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		r, g, b, a := img.At(0, y).RGBA()
		if verbose {
			fmt.Printf("At y %d -> %d,%d,%d,%d\n", y, r, g, b, a)
		}
		if !(a == 0 || a == 65535) || !(r == 0 && g == 0 && b == 0) { // When not black or transparant it's not a valid 9patch
			log.Fatal("Not a valid 9patch file. Wrong pixel at (x,y): (0," + strconv.Itoa(y) + ")")
		}
		if sci.Top == -1 && a == 65535 { // Wait for black pixel
			sci.Top = y - 1
		}
		if sci.Top != -1 && a == 0 { // Wait for transparent pixel
			sci.Bottom = bounds.Max.Y - y - 1
			break
		}
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		r, g, b, a := img.At(x, 0).RGBA()
		if verbose {
			fmt.Printf("At x %d -> %d,%d,%d,%d\n", x, r, g, b, a)
		}
		if !(a == 0 || a == 65535) || !(r == 0 && g == 0 && b == 0) { // When not black or transparant it's not a valid 9patch
			log.Fatal("Not a valid 9patch file. Wrong pixel at (x,y): (" + strconv.Itoa(x) + ",0)")
		}
		if sci.Left == -1 && a == 65535 { // Wait for black pixel
			sci.Left = x - 1
		}
		if sci.Left != -1 && a == 0 { // Wait for transparent pixel
			sci.Right = bounds.Max.X - x - 1
			break
		}
	}
	if verbose {
		fmt.Print(sci, "\n")
	}

	newImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(1, 1, bounds.Max.X-1, bounds.Max.Y-1))

	writer, err := os.Create(filepath.Join(imgDir, sci.Source))
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	err = png.Encode(writer, newImg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Wrote " + sci.Source + " to " + imgDir + "\n")
	writer.Close()

	sciBaseName := imgNoExtBase + ".sci"
	sciWriter, err := os.Create(filepath.Join(imgDir, sciBaseName))
	if err != nil {
		log.Fatal(err)
	}
	defer sciWriter.Close()

	fmt.Fprint(sciWriter, sci)

	if err := sciWriter.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Print("Wrote " + sciBaseName + " to " + imgDir + "\n")
}
