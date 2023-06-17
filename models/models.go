package models

type UserInfoStruct struct {
	UserId          int    `json:"userId"`
	UserEmail       string `json:"useremail"`
	ActivateAccount bool   `json:"activate"`
	Password        string `json:"password"`
}

type MessageError struct {
	Message string `json:"message"`
}

type RefreshTokenStruct struct {
	RefreshToken string `json:"refresh_token"`
}

type CheckUserStruct struct {
	Email string `json:"useremail"`
}

type ChangePassStruct struct {
	OriginPass  string `json:"origin_pass"`
	ConfirmPass string `json:"confirm_pass"`
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

// repositoru structs
type UserInfoResponseStruct struct {
	UserId       int
	UserEmail    string
	Pass         string
	Activate     bool
	RefreshToken string
}

type VisitOnUrl struct {
	Platform string `json:"platform"`
	Count    int    `json:"count"`
}

type InfoUrl struct {
	ShorturlId  int
	OriginalUrl string
}

type UrlsByUserStruct struct {
	OriginUrl string `json:"origin_url"`
	ShortUrl  string `json:"short_url"`
}
