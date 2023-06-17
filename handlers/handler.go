package handlers

import (
	"appurl/models"
	"appurl/repository"
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

const (
	redirectRoute    = "http://127.0.0.1:8001/"
	addActivituRoute = "http://127.0.0.1:8001"
)

type useRepository struct {
	PsqlRepos *repository.InquirysRepository
}

func NewUseRepository(Repo *repository.InquirysRepository) *useRepository {
	return &useRepository{PsqlRepos: Repo}
}

func (rep *useRepository) CreateShortUrl(w http.ResponseWriter, r *http.Request) {
	var (
		originalurl models.UrlReqStruct
		err         error
		message     []byte
		shortedUrl  string
		userEmail   string
		addShortUrl string
	)

	defer func() {
		if err != nil {
			log.Printf("CreateShortUrl(): %s", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	w.Header().Set("Content-Type", "application/json")

	err = json.NewDecoder(r.Body).Decode(&originalurl)
	if err != nil {
		log.Printf("Decode(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error: bad body "})

		return
	}

	userEmail, err = AccessTokenParce(r.Header["Token"][0])
	if err != nil {
		log.Printf("AccessTokenParce(): %s", err.Error())
		message, err = json.Marshal(&models.MessageError{Message: "token invalid"})

		return
	}

	validURL, err := url.ParseRequestURI(originalurl.Url)
	if err != nil {
		log.Printf("ParseRequestURI(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error: invalid URL"})

		return
	}

	shortedUrl, err = rep.PsqlRepos.SelectShortUrl(validURL.String())
	if err != nil {
		log.Printf("SelectShortUrl(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error: invalid URL"})

		return
	}

	if shortedUrl != "" {
		message, err = json.Marshal(&models.ShortUrlRespStruct{ShortUrl: redirectRoute + shortedUrl})
		if err != nil {
			log.Printf("Marshal(): %s", err.Error())
			return
		}
	} else {
		generateUrl, err := GenerationShortUrl(originalurl.Url)
		if err != nil {
			log.Printf("GenerationShortUrl(): %s", err.Error())
			message, _ = json.Marshal(&models.MessageError{Message: "error: invalid url"})

			return
		}

		addShortUrl, err = rep.PsqlRepos.AddGenerateUrl(r.Context(), generateUrl, validURL.String(), userEmail)
		if err != nil {
			log.Printf("AddGenerateUrl(): %s", err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			message, _ = json.Marshal(&models.MessageError{Message: "don't auth"})

			return
		}
		message, _ = json.Marshal(&models.ShortUrlRespStruct{ShortUrl: redirectRoute + addShortUrl})
	}
}

func (rep *useRepository) RedirectShortUrl(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		message   []byte
		userAgent string
		urlInfo   *models.InfoUrl
	)

	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if err != nil {
			log.Printf("RedirectShortUrl(): %s", err)
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write(message)
	}()

	vars := mux.Vars(r)
	index := vars["url_index"]
	userAgent = r.Header["User-Agent"][0]
	remoteAddress := r.RemoteAddr
	res := userAgent + " " + remoteAddress

	urlInfo, err = rep.PsqlRepos.SelectOriginalUrl(index)
	if err != nil {
		log.Printf("SelectOriginalUrl(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error SelectOriginalUrl"})

		return
	}

	_, err = rep.PsqlRepos.AddActivityInfo(r.Context(), index, res, PLatfor(userAgent), urlInfo.ShorturlId)
	if err != nil {
		log.Printf("AddActivityInfo(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error add"})

		return
	}

	http.Redirect(w, r, urlInfo.OriginalUrl, http.StatusFound)
}

func (rep *useRepository) AllUsersUrls(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		message    []byte
		userEmail  string
		userSelect *[]models.UrlsByUserStruct
	)

	w.Header().Set("content-type", "application/json")

	defer func() {
		if err != nil {
			log.Printf("AllUsersUrls(): %s", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	userEmail, err = AccessTokenParce(r.Header["Token"][0])
	if err != nil {
		log.Printf("AccessTokenParce(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error access token"})

		return
	}

	userSelect, err = rep.PsqlRepos.SelectUrlsByUser(r.Context(), userEmail)
	if err != nil {
		log.Printf("SelectUrlsByUser(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "Erorre handlers/handler AllUsersUrls() SelectUrlsByUser()"})

		return
	}

	message, err = json.Marshal(userSelect)
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "Erorre handlers/handler AllUsersUrls()"})

		return
	}
}

func (rep *useRepository) VisitOnUrlH(w http.ResponseWriter, r *http.Request) {
	var (
		message        []byte
		err            error
		visitStatistic []models.VisitOnUrl
	)

	w.Header().Set("content-type", "application/json")

	defer func() {
		if err != nil {
			log.Printf("VisitOnUrlH(): %s", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(message)
	}()

	vars := mux.Vars(r)
	shortURL := vars["url_index"]

	visitStatistic, err = rep.PsqlRepos.VisitStatistic(r.Context(), redirectRoute+shortURL)
	if err != nil {
		log.Printf("VisitStatistic(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "url does not exist "})

		return
	}
	if visitStatistic == nil {
		err = fmt.Errorf("visit statisitc is empty")
		message, _ = json.Marshal(&models.MessageError{Message: "url does not exist "})

		return
	}

	message, err = json.Marshal(visitStatistic)
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "error marshaling "})

		return
	}
}

func (rep *useRepository) CountVisitH(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("CountVisitOnURL(): %s", err.Error())
		message, _ = json.Marshal(&models.MessageError{Message: "search error"})

		return
	}

	message, err = json.Marshal(&models.CountVisit{CountVisit: cV})
	if err != nil {
		log.Printf("Marshal(): %s", err.Error())

		return
	}
}
