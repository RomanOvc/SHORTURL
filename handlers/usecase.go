package handlers

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/speps/go-hashids"
)

type CreateShortUrlResp struct {
	ShortUrl string `json:"shorturl"`
}

func ShortUrlReturn(shorturl string) (*CreateShortUrlResp, error) {
	// load
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	// get
	addr := os.Getenv("DOMAIN")
	if addr == "" {
		return nil, errors.New("missing addres")
	}

	return &CreateShortUrlResp{
		ShortUrl: addr + "/" + shorturl,
	}, err
}

// method create url = 7 symbols
func GenerationShortUrl(s string) (string, error) {
	if s != "" {
		hd := hashids.NewData()
		hd.Salt = s
		hd.MinLength = 7
		h, _ := hashids.NewWithData(hd)
		e, _ := h.EncodeInt64([]int64{1, 2, 3})
		return e, nil
	} else {
		return "string is empty", errors.New("GenerationShortUrl() handler/usecase")
	}
}
