package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	dpi               = flag.Float64("dpi", 200, "screen resolution in Dots Per Inch")
	fontfile          = flag.String("fontfile", "./Monaco-Linux.ttf", "filename of the ttf font")
	hinting           = flag.String("hinting", "none", "none | full")
	size              = flag.Float64("size", 20, "font size in points")
	spacing           = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb              = flag.Bool("whiteonblack", false, "white text on a black background")
	charactersPerLine = flag.Int("charactersPerLine", 15, "The maximum number of characters of text in one line of given image file")
	linesPerImage     = flag.Int("linesPerImage", 5, "Number of lines in an image")
	outputDir         = flag.String("outputDir", "outputs", "The directory where the package will create the output images")
	inputFile         = flag.String("inputFile", "news.txt", "The input file for the source of the text")
	transcriptFile    = flag.String("transcriptFile", "transcript", "The filname for the transcipt file")
)

// ColorPalette is the combination of the background color and foreground color
type ColorPalette struct {
	bgcolor string
	fgcolor string
}

// ImageData is the text being held in a single picture
type ImageData struct {
	textLines    []string
	colorPalette ColorPalette
}

func (imageData ImageData) createImage(fileName string) {
	bgColor, _ := parseHexColor(imageData.colorPalette.bgcolor)
	fgColor, _ := parseHexColor(imageData.colorPalette.fgcolor)

	fg, bg := image.NewUniform(fgColor), image.NewUniform(bgColor)
	rgba := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	// width := 10
	// pt := freetype.Pt(width, width+int(c.PointToFixed(*size)>>6))

	opts := truetype.Options{}
	opts.Size = 125.0
	// face := truetype.NewFace(f, &opts)
	for i, text := range imageData.textLines {
		// awidth, ok := face.GlyphAdvance(rune(text[0]))
		// if ok != true {
		// 	log.Println(err)
		// 	return
		// }
		// iwidthf := int(float64(awidth) / 64)
		pt := freetype.Pt(60, i*220/2+(1080/2-32)-150)
		c.DrawString(text, pt)
	}
	// Save that RGBA image to disk.
	outFile, err := os.Create(*outputDir + "/" + fileName)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Println("Wrote:", fileName)
}

// Color Palettes for the images
var colors = []ColorPalette{
	ColorPalette{bgcolor: "#673ab7", fgcolor: "#ffffff"},
	ColorPalette{bgcolor: "#283593", fgcolor: "#ffffff"},
	ColorPalette{bgcolor: "#9ccc65", fgcolor: "#000000"},
	ColorPalette{bgcolor: "#ffa726", fgcolor: "#000000"},
}

// Takes in a HEX string and returns back a color.RGBA object and an optional error
func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}

// Reads and cleans the file of extra newlines
func readFile(fileName string) string {
	fileBytes, err := ioutil.ReadFile(fileName)
	re := regexp.MustCompile(`([A-Za-z])\n([A-Za-z])`)
	if err != nil {
		fmt.Printf("File not found! %s", fileName)
	}
	newsFileString := string(fileBytes)
	newsFileString = strings.Replace(newsFileString, "\n\n", "", -1)
	newsFileString = re.ReplaceAllString(newsFileString, "{0} {1}")

	return newsFileString
}

// Givem a huge string of text, this will chop it up into appropriate strings of sepcified
// limit length
func chop(content string, delim string, limit int) []string {
	var (
		// A paragraph is a colletion of evenly split sentence
		paragraph []string
		// A sentence is a collection of words, not more than 'count' words
		sentence []string
		// A word is the tuple of a sentence
		words []string
	)

	words = strings.Split(content, delim)

	localCounter := 0

	fmt.Println("Split the content into:", len(words), "words")
	for _, word := range words {
		shouldSkip := false
		if strings.Contains(word, "\n") {
			word = strings.Replace(word, "\n", "", -1)
			shouldSkip = true
		}
		wordLength := len(word)
		if (localCounter+wordLength) >= limit || shouldSkip {
			tempSentence := strings.Join(sentence, " ")
			paragraph = append(paragraph, tempSentence)
			localCounter = 0
			sentence = nil
		}
		localCounter += wordLength
		sentence = append(sentence, word)
	}
	tempSentence := strings.Join(sentence, " ")
	paragraph = append(paragraph, tempSentence)
	return paragraph
}

func main() {
	flag.Parse()
	// var imageData []ImageData
	stringSlice := chop(readFile(*inputFile), " ", 45)

	newpath := filepath.Join(".", *outputDir)
	os.MkdirAll(newpath, os.ModePerm)

	file, _ := os.Create(*transcriptFile)
	defer file.Close()
	for _, sentence := range stringSlice {
		file.WriteString(sentence + "\n---\n")
	}
	randSource := rand.NewSource(time.Now().UnixNano())

	// Pad the slice of strings to match the max line count in an image
	if len(stringSlice)%*linesPerImage != 0 {
		for i := 0; i < *linesPerImage-(len(stringSlice)%*linesPerImage); i++ {
			stringSlice = append(stringSlice, "")
		}
	}

	for i := 0; i < len(stringSlice); i += *linesPerImage {
		var textLines []string
		for j := 0; j < *linesPerImage; j++ {
			fmt.Println(i+j, len(stringSlice))
			if (i + j) >= len(stringSlice) {
				break
			}
			textLines = append(textLines, stringSlice[i+j])
		}
		randomNumber := rand.New(randSource)
		colorPalette := colors[randomNumber.Intn(len(colors))]
		data := ImageData{textLines: textLines, colorPalette: colorPalette}
		fileName := strconv.Itoa(i)
		data.createImage(fileName + ".png")
	}
}
