package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	tmplIndex    = template.Must(template.ParseFiles("template/index.html"))
	tmplCallback = template.Must(template.ParseFiles("template/callback.html"))
	tmplError    = template.Must(template.ParseFiles("template/error.html"))
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p >= 65535 {
		port = "8080"
	}

	goutDebug, _ = strconv.ParseBool(os.Getenv("GOUT_DEBUG"))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.ListenAndServe(":"+port, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmplIndex.Execute(w, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	state := q.Get("state")
	b64Decode, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			"state is not base64 encoded",
			err.Error(),
		})
		return
	}

	split := strings.Split(string(b64Decode), "::")
	if len(split) != 4 {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			"invalid state",
			"invalid state",
		})
		return
	}

	client_id := split[0]
	client_secret := split[1]
	redirect_uri := split[2]
	scope := split[3]
	code := q.Get("code")

	rspToken, err := getToken("common", client_id, client_secret, redirect_uri, scope, code)
	if err != nil {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			"get token error",
			err.Error(),
		})
		return
	}

	if rspToken.fail.Error != "" {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			rspToken.fail.Error,
			rspToken.fail.ErrorDescription + "\nSee: " + rspToken.fail.ErrorUri,
		})
		return
	}

	token := rspToken.success.AccessToken
	rspGetme, err := getMe(token)
	if err != nil {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			"get user error",
			err.Error(),
		})
		return
	}

	if rspGetme.fail.Error.Code != "" {
		tmplError.Execute(w, struct {
			Error       string
			Description string
		}{
			rspGetme.fail.Error.Code,
			rspGetme.fail.Error.Message,
		})
		return
	}

	tmplCallback.Execute(w, struct {
		Email        string
		AccessToken  string
		RefreshToken string
	}{
		rspGetme.success.Mail,
		rspToken.success.AccessToken,
		rspToken.success.RefreshToken,
	})
}
