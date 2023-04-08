package handlers

import (
	"appurl/repository"
	"encoding/json"

	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

type UseRepository struct {
	PsqlRepos *repository.InquirysRepository
}

func NewUseRepository(Repo *repository.InquirysRepository) *UseRepository {
	return &UseRepository{PsqlRepos: Repo}
}

type HandlerInterface interface {
	CreateShortUrl(w http.ResponseWriter, r *http.Request)
}

type UrlReqStruct struct {
	Url string `json:"url"`
}

func (rep *UseRepository) CreateShortUrl(w http.ResponseWriter, r *http.Request) {

	var (
		originalurl   UrlReqStruct
		handlerResult []byte
		err           error
	)

	defer func() {
		if err != nil {
			log.Println(err, "Error request")
			w.WriteHeader(400)
			w.Write(nil)
		} else {
			w.Write(handlerResult)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewDecoder(r.Body).Decode(&originalurl)

	validURL, err := url.ParseRequestURI(originalurl.Url)
	if err != nil {
		return
	}

	generateUrl, err := GenerationShortUrl(validURL.String())
	if err != nil {
		return
	}

	checkUrlDb, err := rep.PsqlRepos.SelectShortUrlCount(generateUrl)
	if err != nil {
		return
	}

	if checkUrlDb == 1 {

		url, err := ShortUrlReturn(generateUrl)
		if err != nil {
			return
		}
		handlerResult, err = json.Marshal(url)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)

	} else {
		rep.PsqlRepos.AddGenerateUrl(generateUrl, originalurl.Url)

		jsonUrl, err := ShortUrlReturn(generateUrl)
		if err != nil {
			return
		}

		handlerResult, err = json.Marshal(jsonUrl)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// Сервис принимает длинный URL, пришедший в POST-запросе.
// Сервис проверят, нет ли уже такого URL в базе данных. Если входящий длинный URL уже существует в системе, то в ней остался и сгенерированный короткий вариант.
// Если длинного URL нет в базе данных или срок действия сокращённой ссылки истёк, необходимо создать новый токен и отправить короткий URL в качестве ответа, сохранив результат в базе данных.
// Сервис отправляет короткий URL в качестве ответа. Статус HTTP 201, если создана новая запись или 200, если запись уже была в базе данных.
func (rep *UseRepository) RedirectShortUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["url_index"]

	checkUrlDb, err := rep.PsqlRepos.SelectOriginalUrl(index)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	} else {
		http.Redirect(w, r, checkUrlDb, http.StatusSeeOther)
	}
}
