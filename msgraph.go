package main

import (
	"github.com/guonaihong/gout"
)

const MS_LOGIN_URL = "https://login.microsoftonline.com/"

var (
	goutClient *gout.Client
	goutDebug  bool
)

type RspBodyToken struct {
	success RspBodyTokenSuccess
	fail    RspBodyTokenFail
}

type RspBodyTokenSuccess struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RspBodyTokenFail struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

type RspBodyGetMe struct {
	success RspBodyGetMeSuccess
	fail    RspBodyError
}

type RspBodyGetMeSuccess struct {
	ID                string `json:"id"`
	Mail              string `json:"mail"`
	DisplayName       string `json:"displayName"`
	UserPrincipalName string `json:"userPrincipalName"`
}

type RspBodyError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func getGoutClient() *gout.Client {
	if goutClient == nil {
		goutClient = gout.NewWithOpt()
	}
	return goutClient
}

func getToken(tenant_id, client_id, client_secret, redirect_uri, scope, code string) (RspBodyToken, error) {
	rsp := RspBodyToken{}

	err := getGoutClient().
		POST(MS_LOGIN_URL + tenant_id + "/oauth2/v2.0/token").
		Debug(goutDebug).
		SetWWWForm(gout.H{
			"client_id":     client_id,
			"client_secret": client_secret,
			"redirect_uri":  redirect_uri,
			"code":          code,
			"scope":         scope,
			"grant_type":    "authorization_code",
		}).
		SetHeader(gout.H{
			"Content-Type": "application/x-www-form-urlencoded",
		}).Callback(func(c *gout.Context) (err error) {
		switch c.Code {
		case 200:
			c.BindJSON(&rsp.success)
		case 400, 401:
			c.BindJSON(&rsp.fail)

		}
		return nil
	}).Do()

	if err != nil {
		return rsp, err
	}

	return rsp, nil
}

func getMe(token string) (RspBodyGetMe, error) {
	rsp := RspBodyGetMe{}

	err := getGoutClient().
		GET("https://graph.microsoft.com/v1.0/me").
		Debug(goutDebug).
		SetHeader(gout.H{
			"Authorization": "Bearer " + token,
		}).Callback(func(c *gout.Context) (err error) {
		switch c.Code {
		case 200:
			c.BindJSON(&rsp.success)
		case 400, 401:
			c.BindJSON(&rsp.fail)
		}
		return nil
	}).Do()

	if err != nil {
		return rsp, err
	}

	return rsp, nil
}
