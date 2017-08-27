package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/parnurzeal/gorequest"
)

var XKCDURL string = "https://xkcd.com/"

//Represents the information for an XKCD comic
type XKCDImage struct {
	Number           string
	Url              string
	OriginalFileName string
	PageUrl          string
}

//Returns file name for comic
func (image *XKCDImage) GetFileName() string {
	return image.Number + "_" + image.OriginalFileName
}

//Returns an XKCDImage by parsing the HTML of an XKCD comic
func getImage(downloadUrl string) (XKCDImage, bool) {
	resp, body, _ := gorequest.New().Get(downloadUrl).End()

	statusCode := strconv.Itoa(resp.StatusCode)
	if string(statusCode[0]) != "2" {
		return XKCDImage{PageUrl: downloadUrl}, true
	}

	var pageLink string
	var imageLink string
	var number string
	var originalFileName string
	var err bool

	for _, line := range strings.Split(body, "\n") {

		if strings.HasPrefix(line, "Permanent link to this comic:") {
			pageLink = line[30:len(line)-6]
			number = pageLink[17:len(pageLink)-1]
		} else if strings.HasPrefix(line, "Image URL (for hotlinking/embedding):") {
			imageLink = line[38:]
			split := strings.Split(imageLink, "/")
			originalFileName = split[len(split)-1]
			if originalFileName == "" {
				err = true
				break
			}
			break
		}
	}

	return XKCDImage{Number: number, Url: imageLink, OriginalFileName: originalFileName, PageUrl: downloadUrl}, err
}

//Downloads an XKCD comic
func downloadImage(image XKCDImage) {
	_, imageBody, _ := gorequest.New().Get(image.Url).End()
	imgBytes := strings.NewReader(imageBody)
	img, _ := os.Create(image.GetFileName())
	defer img.Close()
	io.Copy(img, imgBytes)

}

//Downloads all XKCD Comics
func downloadComics(startComic string) ([]XKCDImage, int) {
	var downloadCount int
	var failedComics []XKCDImage

	currentComicNumber, _ := strconv.Atoi(startComic)
	for ; currentComicNumber != 0; currentComicNumber-- {
		currentURL := XKCDURL + strconv.Itoa(currentComicNumber)
		fmt.Println(currentURL)
		currentImage, err := getImage(currentURL)

		if err {
			fmt.Printf("\t!!!! Failed to download\n")
			failedComics = append(failedComics, currentImage)
		} else {
			fileName := currentImage.GetFileName()
			if _, err := os.Stat(fileName); os.IsNotExist(err) {
				downloadImage(currentImage)
				fmt.Printf("\tDownloaded %v\n", fileName)
				downloadCount++
			} else {
				fmt.Printf("\t%v Already exists. Stopping.\n", fileName)
				break
			}
		}
	}

	return failedComics, downloadCount
}

//Prints failed comics
func printFailed(failedComics *[]XKCDImage) {
	if len(*failedComics) > 0 {
		fmt.Println("\n\nThe following comics failed to download")
		for _, comic := range *failedComics {
			fmt.Printf("\t%v\n", comic.PageUrl)
		}
	}
}

func main() {

	// Get first image to determine start number/url
	firstImage, _ := getImage(XKCDURL)

	failedComics, downloadCount := downloadComics(firstImage.Number)
	printFailed(&failedComics)

	fmt.Printf("\n\nDownloaded %v comics\n", downloadCount)
	fmt.Println("Program Finished")
}
