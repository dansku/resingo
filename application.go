package resingo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

//Application  holds  information about the application that is running on resin.
type Application struct {
	ID         int64  `json:"id"`
	Name       string `json:"app_name"`
	Repository string `json:"git_repository"`
	Metadata   struct {
		URI  string `json:"uri"`
		Type string `json:"type"`
	} `json:"__metadata"`
	DeviceType string `json:"device_type"`
	User       User   `json:"user"`
	Commit     string `json:"commit"`
}

//AppGetAll retrieves all applications that belog to the user in the given
//context.
//
// For this to work, the context should be authorized, probably using the Login
// function.
func AppGetAll(ctx *Context) ([]*Application, error) {
	h := authHeader(ctx.Config.AuthToken)
	uri := ctx.Config.APIEndpoint("application")
	b, err := doJSON(ctx, "GET", uri, h, nil, nil)
	if err != nil {
		return nil, err
	}
	var appRes = struct {
		D []*Application `json:"d"`
	}{}
	err = json.Unmarshal(b, &appRes)
	if err != nil {
		return nil, err
	}
	return appRes.D, nil
}

//AppGetByName returns the application  with the giveb name.
func AppGetByName(ctx *Context, name string) (*Application, error) {
	h := authHeader(ctx.Config.AuthToken)
	uri := ctx.Config.APIEndpoint("application")
	params := make(url.Values)
	params.Set("filter", "app_name")
	params.Set("eq", name)
	b, err := doJSON(ctx, "GET", uri, h, params, nil)
	if err != nil {
		return nil, err
	}
	var appRes = struct {
		D []*Application `json:"d"`
	}{}
	err = json.Unmarshal(b, &appRes)
	if err != nil {
		return nil, err
	}
	if len(appRes.D) > 0 {
		return appRes.D[0], nil
	}
	return nil, errors.New("application not found")
}

//The client expects a 200 status code for a successful reqest, any
// other status code will result into an error with the form [code] [ Any
// content read from the response body]
func do(ctx *Context, method, uri string, header http.Header,
	params url.Values, body io.Reader) ([]byte, error) {
	if params != nil {
		uri = uri + "?" + Encode(params)
	}
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	req.Header = header
	resp, err := ctx.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if !checkStatus(resp.StatusCode) {
		return nil, fmt.Errorf("resingo: [%d ] %s : %s", resp.StatusCode, req.URL.RequestURI(), string(b))
	}
	return b, nil
}

func checkStatus(status int) bool {
	switch status {
	case http.StatusOK, http.StatusCreated:
		return true
	}
	return false
}

func doJSON(ctx *Context, method, uri string, header http.Header,
	params url.Values, body io.Reader) ([]byte, error) {
	header.Set("Content-Type", "application/json")
	return do(ctx, method, uri, header, params, body)
}

//AppGetByID returns application with the given id
func AppGetByID(ctx *Context, id int64) (*Application, error) {
	h := authHeader(ctx.Config.AuthToken)
	uri := ctx.Config.APIEndpoint("application")
	params := make(url.Values)
	params.Set("filter", "id")
	params.Set("eq", fmt.Sprint(id))
	b, err := doJSON(ctx, "GET", uri, h, params, nil)
	if err != nil {
		return nil, err
	}
	var appRes = struct {
		D []*Application `json:"d"`
	}{}
	err = json.Unmarshal(b, &appRes)
	if err != nil {
		return nil, err
	}
	if len(appRes.D) > 0 {
		return appRes.D[0], nil
	}
	return nil, errors.New("application not found")
}

//AppCreate creates a new application with the given name
func AppCreate(ctx *Context, name string, typ DeviceType) (*Application, error) {
	h := authHeader(ctx.Config.AuthToken)
	uri := ctx.Config.APIEndpoint("application")
	data := make(map[string]interface{})
	data["app_name"] = name
	data["device_type"] = typ.String()
	body, err := marhsalReader(data)
	if err != nil {
		return nil, err
	}
	b, err := doJSON(ctx, "POST", uri, h, nil, body)
	if err != nil {
		return nil, err
	}
	rst := &Application{}
	err = json.Unmarshal(b, rst)
	if err != nil {
		return nil, err
	}
	return rst, nil
}

func marhsalReader(o interface{}) (io.Reader, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

//AppDelete removes the application with the given name
func AppDelete(ctx *Context, name string) (bool, error) {
	h := authHeader(ctx.Config.AuthToken)
	uri := ctx.Config.APIEndpoint("application")
	params := make(url.Values)
	params.Set("filter", "app_name")
	params.Set("eq", name)
	b, err := doJSON(ctx, "DELETE", uri, h, params, nil)
	if err != nil {
		return false, err
	}
	return string(b) == "OK", nil
}

//AppGetAPIKey returns the application with the given api key
func AppGetAPIKey(ctx *Context, name string) ([]byte, error) {
	h := authHeader(ctx.Config.AuthToken)
	app, err := AppGetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	end := fmt.Sprintf("application/%d/generate-api-key", app.ID)
	uri := "https://api.resin.io/" + end
	return doJSON(ctx, "POST", uri, h, nil, nil)
}
