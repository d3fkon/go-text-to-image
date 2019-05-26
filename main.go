package main

import (
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
	// "github.com/robfig/graphics-go"
	// "github.com/golang/freetype"
)

var (
	dpi               = flag.Float64("dpi", 200, "screen resolution in Dots Per Inch")
	fontfile          = flag.String("fontfile", "./Monaco-Linux.ttf", "filename of the ttf font")
	hinting           = flag.String("hinting", "none", "none | full")
	size              = flag.Float64("size", 18, "font size in points")
	spacing           = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb              = flag.Bool("whiteonblack", false, "white text on a black background")
	charactersPerLine = flag.Int("charactersPerLine", 45, "The maximum number of characters of text in one line of given image file")
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
	text         string
	colorPalette ColorPalette
}

func (imageData ImageData) createImage(fileName string, wg *sync.WaitGroup) {
	bgColor, _ := parseHexColor(imageData.colorPalette.bgcolor, true)
	fgColor, _ := parseHexColor(imageData.colorPalette.fgcolor, false)

	im, err := gg.LoadJPG("meme.jpg")
	if err != nil {
		log.Fatal(err)
	}
	xSize, ySize := 1920, 1080
	dc := gg.NewContext(xSize, ySize)
	dc.SetRGB(1, 1, 1)
	dc.LoadFontFace("Monaco-Linux.ttf", 30)
	dc.DrawImage(im, 0, 0)

	imageFile, err := os.Open("meme.jpg")
	if err != nil {
	}
	defer imageFile.Close()

	// fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	// f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	ax, ay := 0.0, 0.0
	x, _ := 5.0, 5.0
	maxY := 1900.0
	str := imageData.text
	str = pad(str)
	str = newLines(str)
	strW, strH := dc.MeasureMultilineString(str, 2.0)
	fmt.Println("w, H", strW, strH)

	waterMark := "Read Me Daddy"
	wmW, wmH := dc.MeasureString(waterMark)

	dc.SetColor(bgColor)
	dc.DrawRoundedRectangle(50, float64(ySize)-100, wmW+50, wmH+50, 10)
	dc.DrawRoundedRectangle(float64(xSize/2)-strW/2-30, float64(ySize/2)-130, strW+50, strH+100, 10)
	dc.Fill()
	dc.SetColor(fgColor)
	dc.DrawStringAnchored(waterMark, 80, float64(ySize)-50, 0, 0)
	dc.DrawStringWrapped(str, x, float64(ySize/2)-100, ax, ay, maxY, 2.0, gg.AlignCenter)
	dc.Clip()
	dc.SaveJPG(*outputDir+"/"+fileName, 100)

	fmt.Println("Wrote:", fileName)
	wg.Done()
}
func newLines(str string) string {
	str = strings.Replace(str, "\n", " ", -1)
	linedString := ""
	words := strings.Split(str, " ")
	for i, word := range words {
		if (i+1)%17 == 0 {
			linedString += word + "\n"
		} else {
			linedString += word + " "
		}
	}
	return linedString
}
func pad(str string) string {
	fmt.Println("Str Length:", len(str))
	padStr := ""
	for i := len(str); i < 255; i++ {
		padStr += "-"
	}
	return str + padStr
}

// Color Palettes for the images
var colors = []ColorPalette{
	ColorPalette{bgcolor: "#000000", fgcolor: "#ffffff"},
}

// Takes in a HEX string and returns back a color.RGBA object and an optional error
func parseHexColor(s string, isBg bool) (c color.RGBA, err error) {
	if isBg {
		c.A = 0xCC
	} else {
		c.A = 0xFF
	}
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
	newsFileString = re.ReplaceAllString(newsFileString, "{0} {1}")
	newsFileString = strings.Replace(newsFileString, "\n", "", -1)

	return newsFileString
}

// Givem a huge string of text, this will it up into appropriate strings of sepcified
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

	start := time.Now()
	// var imageData []ImageData
	stringSlice := chop(readFile(*inputFile), " ", 255)

	newpath := filepath.Join(".", *outputDir)
	os.MkdirAll(newpath, os.ModePerm)

	file, _ := os.Create(*transcriptFile)
	defer file.Close()
	for _, sentence := range stringSlice {
		file.WriteString(sentence + "\n---\n")
	}
	randSource := rand.NewSource(time.Now().UnixNano())

	// // Pad the slice of strings to match the max line count in an image
	// if len(stringSlice)%*linesPerImage != 0 {
	// 	for i := 0; i < *linesPerImage-(len(stringSlice)%*linesPerImage); i++ {
	// 		stringSlice = append(stringSlice, "")
	// 	}
	// }
	var wg sync.WaitGroup
	for i := 0; i < len(stringSlice); i += 1 {
		// var textLines []string
		wg.Add(1)
		// for j := 0; j < *linesPerImage; j++ {
		// 	fmt.Println(i+j, len(stringSlice))
		// 	if (i + j) >= len(stringSlice) {
		// 		break
		// 	}
		// 	textLines = append(textLines, stringSlice[i+j])
		// }
		fmt.Println(len(stringSlice))
		randomNumber := rand.New(randSource)
		colorPalette := colors[randomNumber.Intn(len(colors))]
		data := ImageData{text: stringSlice[i], colorPalette: colorPalette}
		fileName := strconv.Itoa(i)
		go data.createImage(fileName+".jpg", &wg)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println(elapsed)
}
