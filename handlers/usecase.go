package handlers

import (
	"fmt"

	"github.com/speps/go-hashids"
)

type createShortUrlResp struct {
	ShortUrl string `json:"shorturl"`
}

func ShortUrlReturn(shorturl string) *createShortUrlResp {
	return &createShortUrlResp{
		ShortUrl: "http://127.0.0.1:8000" + "/" + shorturl,
	}
}

// method create url = 7 symbols
func GenerationShortUrl(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty string")
	}

	hd := hashids.NewData()
	hd.Salt = s
	hd.MinLength = 7

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", fmt.Errorf("hash process: %w", err)
	}

	e, err := h.EncodeInt64([]int64{1, 2, 3})
	if err != nil {
		return "", fmt.Errorf("EncodeInt64: %w", err)
	}

	return e, nil
}

func GenerateResetToken(userEmail string, userId int) (string, error) {
	if userEmail == "" {
		return "", fmt.Errorf("empty user email")
	}

	hd := hashids.NewData()
	hd.Salt = userEmail
	hd.MinLength = 7

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", fmt.Errorf("hash process: %w", err)
	}
	e, err := h.EncodeInt64([]int64{int64(userId)})
	if err != nil {
		return "", fmt.Errorf("EncodeInt64: %w", err)
	}

	return e, nil
}
