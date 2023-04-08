package handlers

import "testing"

type UrlResult struct {
	url    string
	result string
}

func TestGenerateShortUrl(t *testing.T) {
	//just valid url
	datas := []UrlResult{
		{
			url:    "https://www.youtube.com/watch?v=S0Jx6ZjdyO4",
			result: "9pEHRIE",
		},
		{
			url:    "https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package",
			result: "zRVHLIJ",
		},
		{
			url:    "https://www.google.com/search?q=%D0%B3%D0%B5%D0%BD%D0%B5%D1%80%D0%B0%D1%86%D0%B8%D1%8F+%D1%83%D0%BD%D0%B8%D0%BA%D0%B0%D0%BB%D1%8C%D0%BD%D0%BE%D0%B3%D0%BE+7+%D0%B7%D0%BD%D0%B0%D1%87%D0%BD%D0%BE%D0%B3%D0%BE+%D0%B3%D0%BA%D0%B4&sxsrf=APwXEddGwkFLUqFDCCHO613YbSiILASBBw:1680827573881&source=lnms&tbm=isch&sa=X&ved=2ahUKEwjo5uDTwpb-AhVwiYsKHXygB5kQ_AUoAXoECAEQAw&biw=1848&bih=948&dpr=1#imgrc=1RgY9tMSUB9P4M",
			result: "lMgHnIj",
		},
		{
			url:    "",
			result: "string is empty",
		},
	}

	for _, d := range datas {
		got, _ := GenerationShortUrl(d.url)
		want := d.result

		if got != want && len(got) == 7 {
			t.Errorf("got %q, wanted %q", got, want)
		} else {
			t.Errorf("length generate url != 7:%q  generate url len =  %q", len(got), len(want))
		}

	}
}