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

	//"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

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

		partFileName := fmt.Sprintf("%s-part-%d.png", sourceFileName, i)
		//partFileName := fmt.Sprintf("image-part-%d.png", i)
		//partFileName := fmt.Sprintf("image-part-%d-%dx%d.png", i, imageParts[i].Bounds().Max.X, imageParts[i].Bounds().Max.Y)
		outF, err := os.Create(partFileName)
		if err != nil {
			panic(err)
		}
		defer outF.Close()

		err = png.Encode(outF, imageParts[i])
		if err != nil {
			panic(err)
		}

	}

	m := pdf.NewMaroto(consts.Portrait, consts.Letter)
	m.SetPageMargins(20, 10, 20)

	m.Row(250, func() {
		m.Col(12, func() {
			m.FileImage("image-part-0.png", props.Rect{
				Center:  true,
				Percent: 100,
			})
		})
	})

	err = m.OutputFileAndClose("testing.pdf")
	if err != nil {
		fmt.Println("⚠️  Could not save PDF:", err)
		os.Exit(1)
	}
}

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
