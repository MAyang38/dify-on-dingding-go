package models

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type VoicdeRespBody struct {
	CorpusNo string   `json:"corpus_no"`
	ErrMsg   string   `json:"err_msg"`
	ErrNo    int      `json:"err_no"`
	Result   []string `json:"result"`
	Sn       string   `json:"sn"`
}
