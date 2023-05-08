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

// waiting json type
// {
// "url":"path_url"
// }
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
	err = json.NewDecoder(r.Body).Decode(&originalurl)
	if err != nil {
		return
	}

	validURL, err := url.ParseRequestURI(originalurl.Url)
	if err != nil {
		return
	}

	log.Println(validURL.Path)
	generateUrl := GenerationShortUrl(validURL.String())

	checkUrlDb, err := rep.PsqlRepos.SelectShortUrlCount(generateUrl)
	if err != nil {
		return
	}

	if checkUrlDb == 1 {

		url := ShortUrlReturn(generateUrl)

		handlerResult, err = json.Marshal(url)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)

	} else {
		rep.PsqlRepos.AddGenerateUrl(generateUrl, originalurl.Url)

		jsonUrl := ShortUrlReturn(generateUrl)

		handlerResult, err = json.Marshal(jsonUrl)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (rep *UseRepository) RedirectShortUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index := vars["url_index"]

	checkUrlDb, err := rep.PsqlRepos.SelectOriginalUrl(index)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
	} else {
		http.Redirect(w, r, checkUrlDb, http.StatusFound)
	}

}
