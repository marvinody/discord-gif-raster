package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"log"
	"math"
	"os"

	"golang.org/x/image/draw"
)

var (
	horizBlocks int
	vertBlocks  int
	gifFilename string

	discordEmojiBounds image.Rectangle

	discordEmojiVerticalSpacer   = 0
	discordEmojiHorizontalSpacer = 0

	horizontalSpacing int
	verticalSpacing   int
)

func init() {
	flag.IntVar(&horizBlocks, "h", 2, "Minimum horizontal blocks requested")
	flag.IntVar(&vertBlocks, "v", 2, "Minimum vertical blocks requested")
	flag.StringVar(&gifFilename, "f", "", "File to sprint into section")
	flag.Parse()

	discordEmojiBounds = image.Rect(0, 0, 32, 32)
	horizontalSpacing = discordEmojiBounds.Dx() + discordEmojiHorizontalSpacer
	verticalSpacing = discordEmojiBounds.Dy() + discordEmojiVerticalSpacer
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func blocksToPixels(blocks int, isVertical bool) int {
	if isVertical {
		return verticalSpacing * blocks
	}
	return horizontalSpacing * blocks
}

// if not even, will add one block
func pixelsToBlocks(pixels int, isVertical bool) int {
	var blocks float64
	spacing := float64(horizontalSpacing)
	if isVertical {
		spacing = float64(verticalSpacing)
	}
	blocks = float64(pixels) / spacing
	return int(math.Ceil(blocks))
}

// idx -> block for knowing chunk boundaries
// follows standard left->right and then top->down with increasing idx
func indexToBlocks(idx int) (int, int) {
	return idx % horizBlocks, idx / horizBlocks
}

func main() {
	if gifFilename == "" {
		log.Fatal("Please pass a file with the -f flag")
	}
	f, err := os.Open(gifFilename)
	if err != nil {
		log.Fatalln(err)
	}
	img, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	//newWidth := horizBlocks*32 + (horizBlocks-1)*5
	//newHeight := vertBlocks*32 + (vertBlocks-1)*3
	newWidth := blocksToPixels(horizBlocks, false)
	newHeight := blocksToPixels(vertBlocks, true)

	origWidth, origHeight := img.Config.Width, img.Config.Height
	AR := float64(origWidth) / float64(origHeight)
	// if widths are closers than heights
	widthRatio := float64(origWidth) / float64(newWidth)
	heightRatio := float64(origHeight) / float64(newHeight)
	if widthRatio < heightRatio {
		// hardset the width, change height
		newHeight = int(float64(newWidth) / AR)
		vertBlocks = pixelsToBlocks(newHeight, true)
	} else { // heights are closer
		// hardset the height, and change width
		newWidth = int(AR * float64(newHeight))
		horizBlocks = pixelsToBlocks(newWidth, false)
	}
	// scaler is constant between all frames because same output sizes
	scaler := draw.BiLinear.NewScaler(newWidth, newHeight, origWidth, origHeight)

	resizedBounds := image.Rect(0, 0, newWidth, newHeight)

	// create segmented gifs
	segmentedGifs := make([]*gif.GIF, horizBlocks*vertBlocks)
	for idx := range segmentedGifs {
		segmentedGifs[idx] = &gif.GIF{
			// will be filled in later
			Image:     make([]*image.Paletted, len(img.Image)),
			Delay:     img.Delay,
			LoopCount: img.LoopCount,
			Disposal:  img.Disposal,
			Config: image.Config{
				ColorModel: img.Config.ColorModel,
				Width:      discordEmojiBounds.Dx(),
				Height:     discordEmojiBounds.Dy(),
			},
			BackgroundIndex: img.BackgroundIndex,
		}
	}

	for idx, frame := range img.Image {
		// resized frame
		scaled := image.NewPaletted(resizedBounds, frame.Palette)
		scaledImageBounds := scaled.Bounds()
		scaler.Scale(scaled, scaledImageBounds, frame, frame.Bounds(), draw.Src, nil)
		// now crop the frame for each block
		for segmentIdx, segmentedGif := range segmentedGifs {
			cropped := image.NewPaletted(discordEmojiBounds, frame.Palette)
			// find block and coords
			// find where we start in the source frame
			blockX, blockY := indexToBlocks(segmentIdx)
			startX, startY := blocksToPixels(blockX, false), blocksToPixels(blockY, true)
			cropCornerStart := image.Pt(startX, startY)
			// what chunk are we taking from source?
			croppingBounds := cropped.Bounds()
			// some bounds checking
			endBounds := croppingBounds.Add(cropCornerStart)
			// so, if we out "cropping" region is OOB, limit it
			if endBounds.Max.X > scaledImageBounds.Max.X {
				croppingBounds.Max.X = scaledImageBounds.Max.X - endBounds.Min.X
			}
			if endBounds.Max.Y > scaledImageBounds.Max.Y {
				croppingBounds.Max.Y = scaledImageBounds.Max.Y - endBounds.Min.Y
			}

			// put transparent background before copying main data
			draw.Draw(cropped, cropped.Bounds(), image.Transparent, image.ZP, draw.Src)

			// copy from resized frame to cropped frame
			draw.Draw(cropped, croppingBounds, scaled, cropCornerStart, draw.Src)
			segmentedGif.Image[idx] = cropped
		}

	}

	os.Mkdir("gifs", 0700)
	for i := 0; i < horizBlocks*vertBlocks; i++ {
		f, _ := os.Create(fmt.Sprintf("gifs/%03d.gif", i))
		gif.EncodeAll(f, segmentedGifs[i])
	}
	fmt.Println("Upload all the gifs in gifs directory then copy paste the below")
	idx := 0
	for y := 0; y < vertBlocks; y++ {
		for x := 0; x < horizBlocks; x++ {
			fmt.Printf(":%03d: ", idx)
			idx++
		}
		fmt.Printf("\n")
	}
}

func findLeastUsedColor(p image.PalettedImage, bounds image.Rectangle) uint8 {
	hits := map[uint8]int{}
	max := bounds.Max
	min := bounds.Min
	for y := min.Y; y < max.Y; y++ {
		for x := min.X; x < max.X; x++ {
			idx := p.ColorIndexAt(x, y)
			hits[idx]++
		}
	}
	minKey, minVal := 0, math.MaxInt32
	for i := 0; i < math.MaxUint8; i++ {
		if hits[uint8(i)] < minVal {
			minKey = i
			minVal = hits[uint8(i)]
		}
	}
	return uint8(minKey)
}
