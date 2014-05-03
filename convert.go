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
	return "border.left: " + strconv.Itoa(self.Left) + "\nborder.right: " + strconv.Itoa(self.Right) + "\nborder.bottom: " + strconv.Itoa(self.Bottom) + "\nborder.right: " + strconv.Itoa(self.Right) + "\nsource: " + self.Source
}

func main() {
	var sImg string
	var verbose bool
	flag.StringVar(&sImg, "img", "", "The 9patch image")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()

	sPathImg := filepath.Dir(sImg)
	sBaseImg := filepath.Base(sImg)

	if !strings.HasSuffix(sBaseImg, ".9.png") {
		log.Fatal("Provided image doesn't have .9.png extention")
	}

	sBaseNoExtImg := sBaseImg[0 : len(sBaseImg)-6]
	if verbose {
		fmt.Print("Image Path: " + sPathImg + " Base " + sBaseImg + "\n")
	}

	reader, err := os.Open(sImg)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	reader.Close()

	bounds := m.Bounds()
	sci := Sci{-1, -1, -1, -1, sBaseNoExtImg + ".png"}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		r, g, b, a := m.At(0, y).RGBA()
		if verbose {
			fmt.Printf("At y %d -> %d,%d,%d,%d\n", y, r, g, b, a)
		}
		if !(a == 0 || a == 65535) || !(r == 0 && g == 0 && b == 0) {
			log.Fatal("Not a valid 9patch file. Wrong pixel at (x,y): (0," + strconv.Itoa(y) + ")")
		}
		if sci.Top == -1 && a == 65535 {
			sci.Top = y - 1
		}
		if sci.Top != -1 && a == 0 {
			sci.Bottom = bounds.Max.Y - y - 1
			break
		}
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		r, g, b, a := m.At(x, 0).RGBA()
		if verbose {
			fmt.Printf("At x %d -> %d,%d,%d,%d\n", x, r, g, b, a)
		}
		if !(a == 0 || a == 65535) || !(r == 0 && g == 0 && b == 0) {
			log.Fatal("Not a valid 9patch file. Wrong pixel at (x,y): (" + strconv.Itoa(x) + ",0)")
		}
		if sci.Left == -1 && a == 65535 {
			sci.Left = x - 1
		}
		if sci.Left != -1 && a == 0 {
			sci.Right = bounds.Max.X - x - 1
			break
		}
	}
	if verbose {
		fmt.Print(sci, "\n")
	}

	newM := m.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(1, 1, bounds.Max.X-1, bounds.Max.Y-1))

	writer, err := os.Create(filepath.Join(sPathImg, sci.Source))
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	err = png.Encode(writer, newM)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Wrote " + sci.Source + " to " + sPathImg + "\n")
	writer.Close()

	sciBaseName := sBaseNoExtImg + ".sci"
	sciWriter, err := os.Create(filepath.Join(sPathImg, sciBaseName))
	if err != nil {
		log.Fatal(err)
	}
	defer sciWriter.Close()

	fmt.Fprint(sciWriter, sci)

	if err := sciWriter.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Print("Wrote " + sciBaseName + " to " + sPathImg + "\n")
}
