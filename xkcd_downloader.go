package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/parnurzeal/gorequest"
)

var XKCDURL string = "https://xkcd.com/"
var IMAGEDIR string = "images"

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
func getImage(downloadUrl string) (image XKCDImage, err bool) {
	image.PageUrl = downloadUrl
	resp, body, _ := gorequest.New().Get(downloadUrl).End()

	statusCode := strconv.Itoa(resp.StatusCode)
	if string(statusCode[0]) != "2" {
		err = true
		return image, err
	}

	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "Permanent link to this comic:") {
			var pageLink string = line[30:len(line)-6]
			image.Number = pageLink[17:len(pageLink)-1]
		} else if strings.HasPrefix(line, "Image URL (for hotlinking/embedding):") {
			var imageLink string = line[38:]
			image.Url = imageLink

			split := strings.Split(imageLink, "/")
			var originalFileName string = split[len(split)-1]
			image.OriginalFileName = originalFileName

			if originalFileName == "" {
				err = true
			}
			break
		}
	}

	return image, err
}

//Downloads an XKCD comic
func downloadImage(image XKCDImage) {
	_, imageBody, _ := gorequest.New().Get(image.Url).End()
	imgBytes := strings.NewReader(imageBody)
	img, _ := os.Create(getDownloadPath(image))
	defer img.Close()
	io.Copy(img, imgBytes)
}

func downloadSingleComic(currentComicNumber int, failedComics *[]XKCDImage, c chan int) {
	currentURL := XKCDURL + strconv.Itoa(currentComicNumber)
	currentImage, err := getImage(currentURL)

	if err {
		fmt.Printf("%v\n\t!!!! Failed to download\n", currentURL)
		*failedComics = append(*failedComics, currentImage)
		c <- 3
	} else {
		fileName := currentImage.GetFileName()
		if _, err := os.Stat(getDownloadPath(currentImage)); os.IsNotExist(err) {
			downloadImage(currentImage)
			fmt.Printf("%v\n\tDownloaded %v\n", currentURL, fileName)
			c <- 0
		} else {
			fmt.Printf("%v\n\t%v Already exists. Stopping.\n", currentURL, fileName)
			c <- 1
		}
	}
}

//Downloads all XKCD Comics
func downloadComics(startComic string) (failedComics []XKCDImage, downloadCount int) {
	currentComicNumber, _ := strconv.Atoi(startComic)

	downloadChannel := make(chan int)
	for ;currentComicNumber > 0; {
		go downloadSingleComic(currentComicNumber, &failedComics, downloadChannel)
		currentComicNumber--
		go downloadSingleComic(currentComicNumber, &failedComics, downloadChannel)
		currentComicNumber--
		go downloadSingleComic(currentComicNumber, &failedComics, downloadChannel)
		currentComicNumber--
		go downloadSingleComic(currentComicNumber, &failedComics, downloadChannel)
		currentComicNumber--
		go downloadSingleComic(currentComicNumber, &failedComics, downloadChannel)
		currentComicNumber--
		a, b, c, d ,e := <-downloadChannel,<-downloadChannel,<-downloadChannel,<-downloadChannel,<-downloadChannel

		if a == 1 || b == 1 || c == 1 || d == 1 || e == 1 {
			break
		}
	}

	return failedComics, downloadCount
}

//Return path to save image
func getDownloadPath(image XKCDImage) string {
	return filepath.Join(IMAGEDIR, image.GetFileName())
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

//Create directory to save images
func createImageDir() {
	os.Mkdir(IMAGEDIR, os.ModeDir)
}

func main() {
	createImageDir()

	// Get first image to determine start number/url
	firstImage, _ := getImage(XKCDURL)

	failedComics, downloadCount := downloadComics(firstImage.Number)
	printFailed(&failedComics)

	fmt.Printf("\n\nDownloaded %v comics\n", downloadCount)
	fmt.Println("Program Finished")
}
