package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

const DPI = 300

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Source image path: ")
	sourceImagePath, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	sourceImagePath, err = filepath.Abs(strings.Trim(sourceImagePath, "\r\n"))

	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", sourceImagePath)

	sourceFileName := filepath.Base(sourceImagePath)
	fmt.Println(sourceFileName)

	openedImage, err := os.Open(sourceImagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer openedImage.Close()

	imgData, err := png.Decode(openedImage)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bounds: ", imgData.Bounds())
	//fmt.Println(imgData.At(0, 0).RGBA())

	imageRects := CreateImageRectangles(imgData.Bounds())

	fmt.Println(imageRects)

	imageParts := make([]*image.RGBA, 0)
	//imageParts[0].Set()

	// Create blank images for each rectangle
	for _, imageRect := range imageRects {
		imageParts = append(imageParts, image.NewRGBA(imageRect))
	}

	//create new image with dim. 7.5 x 10in
	//imageParts = append(imageParts, image.NewRGBA(image.Rect(0, 0, 2250, 3000)))

	edgeLineSize := 2

	// Set up PDF object
	pdf := gofpdf.New("P", "in", "Letter", "")

	for i := range imageParts {
		//loop though image and copy everything into first section
		for y := imageParts[i].Rect.Min.Y; y < imageParts[i].Rect.Max.Y; y++ {
			for x := imageParts[i].Rect.Min.X; x < imageParts[i].Rect.Max.X; x++ {
				imageParts[i].Set(x, y, imgData.At(x, y))

				//if on a boarder, set black outline
				if x <= imageParts[i].Rect.Min.X+edgeLineSize || x >= imageParts[i].Rect.Max.X-edgeLineSize || y <= imageParts[i].Rect.Min.Y+edgeLineSize || y >= imageParts[i].Rect.Max.Y-edgeLineSize {
					imageParts[i].Set(x, y, color.Black)

				}
			}
		}

		// Output Image File
		partFileName := fmt.Sprintf("%s-part-%d.png", sourceFileName, i)
		//partFileName := fmt.Sprintf("image-part-%d.png", i)

		outF, err := os.Create(partFileName)
		if err != nil {
			panic(err)
		}
		defer outF.Close()

		err = png.Encode(outF, imageParts[i])
		if err != nil {
			panic(err)
		}

		// Add image to pdf
		AddImageToPdfPage(pdf, partFileName, imageParts[i].Bounds(), DPI)
	}

	// Output pdf to a file
	fileStr := "Output.pdf"
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		panic(err)
	}

}

// Adds page to pdf. Pass reference to pdf file object
func AddImageToPdfPage(pdf *gofpdf.Fpdf, pngFile string, bounds image.Rectangle, dpi int) {
	// Set margins
	mx := 0.5
	my := 0.5

	var opt gofpdf.ImageOptions
	pdf.AddPage()
	pdf.SetFont("Arial", "", 11)
	pdf.SetX(0)
	opt.ImageType = "png"
	// Convert width and height to inches
	width := float64(bounds.Dx()) / float64(dpi)
	height := float64(bounds.Dy()) / float64(dpi)

	pdf.ImageOptions(pngFile, mx, my, width, height, false, opt, 0, "")
	opt.AllowNegativePosition = true
}

// Pass the bounds of the rectangle
// Returns an array of rectangles by splitting the image into 8.5x11 pages
func CreateImageRectangles(imgBounds image.Rectangle) []image.Rectangle {
	const maxPageX = 2250
	const maxPageY = 3000

	imageRectangles := make([]image.Rectangle, 0)

	imgSize := imgBounds.Size()
	//if image is within 7.5 x 10in
	if imgSize.X < maxPageX && imgSize.Y < maxPageY {
		imageRectangles = append(imageRectangles, imgBounds)
		return imageRectangles[:]
	}

	extraX := imgSize.X
	extraY := imgSize.Y

	// While there is overflowing size, keep looping and creating rectangles
	baseX := 0
	baseY := 0
	targetX := 0
	targetY := 0
	for extraY > 0 {
		if extraY > maxPageY {
			targetY = maxPageY
		} else {
			targetY = extraY
		}

		//fmt.Printf("base: %d - extra: %d - target: %d\n", baseX, extraX, targetX)
		for extraX > 0 {
			if extraX > maxPageX {
				targetX = maxPageX
			} else {
				targetX = extraX
			}

			imageRectangles = append(imageRectangles, image.Rect(baseX, baseY, baseX+targetX, baseY+targetY))

			baseX += targetX
			extraX -= targetX
			//fmt.Printf("base: %d - extra: %d - target: %d\n", baseX, extraX, targetX)
			//fmt.Println(imageRectangles)
		}
		baseY += targetY
		extraY -= targetY

		//reset X variables
		baseX = 0
		extraX = imgSize.X
		targetX = 0
	}

	return imageRectangles[:]
}
