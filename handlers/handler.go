package handlers

import (
	"appurl/repository"
	"encoding/json"

	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

const (
	redirectRoute    = "http://127.0.0.1:8001/"
	addActivituRoute = "http://127.0.0.1:8001"
)

type UseRepository struct {
	PsqlRepos *repository.InquirysRepository
}

func NewUseRepository(Repo *repository.InquirysRepository) *UseRepository {
	return &UseRepository{PsqlRepos: Repo}
}

type UrlReqStruct struct {
	Url string `json:"url"`
}
type ShortUrlRespStruct struct {
	ShortUrl string `json:"short_url"`
}

type AllUsersUrlsStruct struct {
	OriginalUrl string `json:"origin_url"`
	ShortUrl    string `json:"short_url"`
}

type CountVisit struct {
	CountVisit int `json:"count_visit"`
}

func (rep *UseRepository) CreateShortUrl(w http.ResponseWriter, r *http.Request) {

	var (
		originalurl UrlReqStruct
		err         error
		message     []byte
		shortedUrl  string
		userEmail   string
		addShortUrl string
	)

	defer func() {
		if err != nil {
			log.Println(err, "error request handlers/handler CreateShortUrl()")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(message)
		} else {
			w.Write(message)
		}
	}()

	w.Header().Set("Content-Type", "application/json")

	userEmail, err = AccessTokenParce(r.Header["Token"][0])
	if err != nil {
		log.Println("token invalid")

		message, err = json.Marshal(&MessageError{Message: "token invalid"})

		return
	}

	err = json.NewDecoder(r.Body).Decode(&originalurl)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error: bad body "})
		log.Println("body error")

		return
	}

	validURL, err := url.ParseRequestURI(originalurl.Url)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error: invalid URL"})
		log.Println("invalid url handlers/handler CreateShortUrl() method ParseRequestURI()")

		return
	}

	shortedUrl, err = rep.PsqlRepos.SelectShortUrl(validURL.String())
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error: invalid URL"})
		log.Println("invalid url handlers/handler CreateShortUrl() method SelectShortUrl()")

		return
	}

	if shortedUrl != "" {
		message, err = json.Marshal(&ShortUrlRespStruct{ShortUrl: redirectRoute + shortedUrl})
		if err != nil {
			log.Println("error Marshal shorturl")

			return
		}
	} else {
		generateUrl, err := GenerationShortUrl(originalurl.Url)
		if err != nil {
			message, _ = json.Marshal(&MessageError{Message: "error: invalid url"})
			log.Println("error: handlers/handler CreateShortUrl() GenerationShortURl()")

			return
		}

		addShortUrl, err = rep.PsqlRepos.AddGenerateUrl(r.Context(), generateUrl, validURL.String(), userEmail)
		if err != nil {
			log.Println("error AddGenerateUrl()")
			w.WriteHeader(http.StatusUnauthorized)
			message, _ = json.Marshal(&MessageError{Message: "don't auth"})

			return
		}
		message, _ = json.Marshal(&ShortUrlRespStruct{ShortUrl: redirectRoute + addShortUrl})
	}
}

func (rep *UseRepository) RedirectShortUrl(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		message    []byte
		userAgent  string
		activityId int
		urlInfo    *repository.InfoUrl
	)

	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(message)
		}
	}()

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	index := vars["url_index"]
	userAgent = r.Header["User-Agent"][0]
	remoteAddress := r.RemoteAddr
	res := userAgent + " " + remoteAddress

	urlInfo, err = rep.PsqlRepos.SelectOriginalUrl(index)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error SelectOriginalUrl"})
		log.Println("error handlers/handler RedirectShortUrl() SelectOriginUrl()")

		return
	}

	// log.Println(urlInfo)

	userPlatform := PLatfor(userAgent)

	activityId, err = rep.PsqlRepos.AddActivityInfo(r.Context(), index, res, userPlatform, urlInfo.ShorturlId)
	if err != nil {
		message, _ = json.Marshal(&MessageError{Message: "error add"})
		log.Println("error handlers/handler RedirectShortUrl() AddActivityInfo()")

		return
	}
	log.Println(activityId)
	http.Redirect(w, r, urlInfo.OriginalUrl, http.StatusFound)
}

func (rep *UseRepository) AllUsersUrls(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		message    []byte
		userEmail  string
		userSelect *[]repository.UrlsByUserStruct
	)

	defer func() {
		if err != nil {
			w.Write(message)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Write(message)
		}
	}()

	w.Header().Set("content-type", "application/json")

	userEmail, err = AccessTokenParce(r.Header["Token"][0])
	if err != nil {
		log.Println("error: handlers/handler AllUsersUrls() AccessTokenParce() ")
		message, _ = json.Marshal(&MessageError{Message: "error access token"})

		return
	}

	userSelect, err = rep.PsqlRepos.SelectUrlsByUser(r.Context(), userEmail)
	if err != nil {
		log.Println("error Marshal")
		message, _ = json.Marshal(&MessageError{Message: "Erorre handlers/handler AllUsersUrls() SelectUrlsByUser()"})

		return
	}

	message, err = json.Marshal(userSelect)
	if err != nil {
		log.Println("error Marshal")
		message, _ = json.Marshal(&MessageError{Message: "Erorre handlers/handler AllUsersUrls()"})

		return
	}
}

func (rep *UseRepository) VisitOnUrlH(w http.ResponseWriter, r *http.Request) {
	var (
		message        []byte
		err            error
		visitStatistic []repository.VisitOnUrl
	)

	defer func() {
		if err != nil {
			log.Println(err, "error request handlers/handler CreateShortUrl()")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(message)
		} else {
			w.Write(message)
		}
	}()

	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	shortURL := vars["url_index"]
	fullPath := redirectRoute + shortURL

	visitStatistic, err = rep.PsqlRepos.VisitStatistic(r.Context(), fullPath)
	if err != nil {
		log.Println("err handlers/handler in VisitOnUrlH() method VisitStatistic()")
		message, _ = json.Marshal(&MessageError{Message: "url does not exist "})

		return
	}
	if visitStatistic == nil {
		log.Println("err handlers/handler in VisitOnUrlH() method VisitStatistic()")
		message, _ = json.Marshal(&MessageError{Message: "url does not exist "})

		return
	}

	message, err = json.Marshal(visitStatistic)
	if err != nil {
		log.Println("err handlers/handler in VisitOnUrlH() method VisitStatistic()")
		message, _ = json.Marshal(&MessageError{Message: "error marshaling "})

		return
	}
}

func (rep *UseRepository) CountVisitH(w http.ResponseWriter, r *http.Request) {
	var (
		message []byte
		err     error
	)

	w.Header().Set("content-type", "application/json")

	defer func() {
		if err != nil {
			w.Write(message)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Write(message)
		}
	}()

	res := mux.Vars(r)
	index := res["url_index"]
	url := redirectRoute + index

	cV, err := rep.PsqlRepos.CountVisitOnURL(r.Context(), url)
	if err != nil {
		log.Println("err handlers/handler in VisitOnUrlH() method VisitStatistic()")
		message, _ = json.Marshal(&MessageError{Message: "search error"})

		return
	}

	message, err = json.Marshal(&CountVisit{CountVisit: cV})
	if err != nil {
		log.Println("error Marshaling")

		return
	}
}
