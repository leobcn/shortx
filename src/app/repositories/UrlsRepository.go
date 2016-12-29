package repositories

import (
	"app/models"
	"app/utils"
	"configs"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jinzhu/gorm"
)

// Urls repository struct
type UrlsRepository struct {
	//
}

// Get long urls from given data
func (urlsRepository *UrlsRepository) GetLongUrls(body io.ReadCloser) map[string]string {

	decoder := json.NewDecoder(body)

	var urls map[string]map[string]string

	err := decoder.Decode(&urls)

	if err != nil {
		panic(err)
	}

	return urls["long_urls"]
}

// Generate short urls from the given long urls
func (urlsRepository *UrlsRepository) GenerateShortUrls(Urls map[string]string, request *http.Request) map[string]string {
	db := DbRepository{}.init()

	for longUrl, _ := range Urls {
		// channels
		shortUrlExist := make(chan bool)
		shortUrl := make(chan string)

		go urlsRepository.shortUrlExist(db, longUrl, shortUrlExist, shortUrl)

		if <-shortUrlExist {
			Urls[longUrl] = "http://" + request.Host + "/" + <-shortUrl
		} else {
			Urls[longUrl] = "http://" + request.Host + "/" + urlsRepository.generateShortUrl(db, longUrl)
		}
	}

	defer db.Close()

	return Urls
}

// Generate short url from the given long url
func (urlsRepository *UrlsRepository) generateShortUrl(db *gorm.DB, longUrl string) string {
	for {
		randomString := utils.RandomString(configs.SHORT_URL_STRING_SIZE)

		if db.Where("short_url = ?", randomString).First(&models.Url{}).Error != nil {
			url := models.Url{LongUrl: longUrl, ShortUrl: randomString}
			db.Create(&url)

			return randomString
		}
	}
}

// Determine short url existence
func (urlsRepository *UrlsRepository) shortUrlExist(db *gorm.DB, longUrl string, shortUrlExist chan bool, shortUrl chan string) {
	var url models.Url
	query := db.Where("long_url = ?", longUrl).First(&url)

	if query.Error == nil {
		shortUrlExist <- true
		shortUrl <- url.ShortUrl
	}

	shortUrlExist <- false
	shortUrl <- ""
}

// Get long url from the given short url
func (urlsRepository *UrlsRepository) GetLongUrl(shortUrl string) (bool, string) {
	db := DbRepository{}.init()
	var url models.Url

	query := db.Where("short_url = ?", shortUrl).First(&url)

	defer db.Close()

	if query.Error == nil {
		return true, url.LongUrl
	}

	return false, ""
}
