package handlers

import (
	"github.com/speps/go-hashids"
)

type CreateShortUrlResp struct {
	ShortUrl string `json:"shorturl"`
}

func ShortUrlReturn(shorturl string) *CreateShortUrlResp {
	return &CreateShortUrlResp{
		ShortUrl: "http://127.0.0.1:8000" + "/" + shorturl,
	}
}

// method create url = 7 symbols
func GenerationShortUrl(s string) string {
	if s != "" {
		hd := hashids.NewData()
		hd.Salt = s
		hd.MinLength = 7
		h, _ := hashids.NewWithData(hd)
		e, _ := h.EncodeInt64([]int64{1, 2, 3})
		return e
	}
	return ""
}
