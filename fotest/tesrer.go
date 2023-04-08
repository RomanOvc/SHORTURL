package main

import (
	"errors"
	"fmt"
)

// func GenerationShortUrl(s string) string {
// 	if s != "" {
// 		hd := hashids.NewData()
// 		hd.Salt = s
// 		hd.MinLength = 7
// 		h, _ := hashids.NewWithData(hd)
// 		e, _ := h.EncodeInt64([]int64{1, 2, 3})
// 		return e
// 	} else {
// 		return "string is empty"
// 	}

// }

type myError struct {
	text string
	desc string
}

func (my *myError) Error() string {
	return my.text + ":" + my.desc
}

func Wrap(err error, text string) error {
	return fmt.Errorf("%s: %w", text, err)
}

func newError(m string) error {
	return &myError{
		text: m,
		desc: "custom error",
	}
}

func main() {
	// fmt.Println(GenerationShortUrl("https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package"))
	// fmt.Println(GenerationShortUrl("https://www.youtube.com/watch?v=S0Jx6ZjdyO4"))
	// fmt.Println(GenerationShortUrl("https://www.google.com/search?q=%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%86%D0%B8%D1%8F+%D1%83%D0%BD%D0%B8%D0%BA%D0%B0%D0%BB%D1%8C%D0%BD%D0%BE%D0%B3%D0%BE+7+%D0%B7%D0%BD%D0%B0%D1%87%D0%BD%D0%BE%D0%B3%D0%BE+%D0%B3%D0%BA%D0%B4&sxsrf=APwXEddGwkFLUqFDCCHO613YbSiILASBBw:1680827573881&source=lnms&tbm=isch&sa=X&ved=2ahUKEwjo5uDTwpb-AhVwiYsKHXygB5kQ_AUoAXoECAEQAw&biw=1848&bih=948&dpr=1#imgrc=1RgY9tMSUB9P4M"))
	// fmt.Println(GenerationShortUrl(";lasdka;sldkasl;d"))

	// err := errors.New("123214")

	// err1 := fmt.Errorf("error")
	err1 := newError("my error")

	err := fmt.Errorf("annotation: %w", err1)

	fmt.Println(errors.Is(err, err1))
	fmt.Println(err)

}
