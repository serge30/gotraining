package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type GiphyImageData struct {
	Url    string
	Width  string
	Height string
}

type GiphyImages struct {
	Original GiphyImageData
}

type GiphyObject struct {
	Slug   string
	Url    string
	Images GiphyImages
	Title  string
}

type GiphyTrendingResponse struct {
	Data []GiphyObject
}

const GIPHY_TRENDING_URL = "https://api.giphy.com/v1/gifs/trending"

var (
	apiKey          = flag.String("api_key", "", "Giphy API key")
	limit           = flag.Int("limit", -1, "Max number of items to return")
	outputDirectory = flag.String("output_dir", "", "Directory to save downloaded GIFs")
)

func GetGiphyTrending(apiKey string, limit int) (GiphyTrendingResponse, error) {
	url := fmt.Sprintf("%s?api_key=%s", GIPHY_TRENDING_URL, apiKey)

	if limit > 0 {
		url = fmt.Sprintf("%s&limit=%d", url, limit)
	}
	response, err := http.Get(url)
	if err != nil {
		return GiphyTrendingResponse{}, err
	}

	if response.StatusCode != 200 {
		return GiphyTrendingResponse{}, fmt.Errorf("Giphy returned HTTP status: %s", response.Status)
	}

	var responseObj GiphyTrendingResponse

	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&responseObj)
	if err != nil {
		return GiphyTrendingResponse{}, err
	}

	return responseObj, nil
}

func DownloadFile(fileUrl string, outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	response, err := http.Get(fileUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadItem(item GiphyObject, outputDir string) (string, error) {
	filePath, err := filepath.Abs(filepath.Join(outputDir, item.Slug+".gif"))
	if err != nil {
		return "", err
	}

	err = DownloadFile(item.Images.Original.Url, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func main() {
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("Giphy API key is not specified")
	}

	responseObj, err := GetGiphyTrending(*apiKey, *limit)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range responseObj.Data {
		fmt.Printf("%s (%s)\n", item.Title, item.Url)
		if item.Images.Original.Url != "" {
			filePath, err := DownloadItem(item, *outputDirectory)
			if err != nil {
				log.Println(err)
			} else {
				fmt.Printf("saved to %s\n", filePath)
			}
		}
	}
}
