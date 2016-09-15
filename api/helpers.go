// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	result, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"reason":  message,
	})
	return c.String(status, string(result))
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	payload["success"] = true
	result, _ := json.Marshal(payload)
	return c.String(http.StatusOK, string(result))
}

//LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c echo.Context, l zap.Logger) error {
	l.Debug("Loading payload...")

	data, err := GetRequestBody(c)
	if err != nil {
		l.Error("Loading payload failed.", zap.Error(err))
		return err
	}

	err = json.Unmarshal([]byte(data), payloadStruct)
	if err != nil {
		l.Error("Loading payload failed.", zap.Error(err))
		return err
	}

	var jsonPayload map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonPayload)
	if err != nil {
		l.Error("Loading payload failed.", zap.Error(err))
		return err
	}

	var missingFieldErrors []string
	v := reflect.ValueOf(payloadStruct).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r, n := utf8.DecodeRuneInString(t.Field(i).Name)
		field := string(unicode.ToLower(r)) + t.Field(i).Name[n:]
		if jsonPayload[field] == nil {
			missingFieldErrors = append(missingFieldErrors, fmt.Sprintf("%s is required", field))
		}
	}

	if len(missingFieldErrors) != 0 {
		err := errors.New(strings.Join(missingFieldErrors[:], ", "))
		l.Error("Loading payload failed.", zap.Error(err))
		return err
	}

	l.Debug("Payload loaded successfully.")
	return nil
}

//GetRequestBody from echo context
func GetRequestBody(c echo.Context) (string, error) {
	bodyCache := c.Get("requestBody")
	if bodyCache != nil {
		return bodyCache.(string), nil
	}
	body := c.Request().Body()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	c.Set("requestBody", string(b))
	return string(b), nil
}

//GetRequestJSON as the specified interface from echo context
func GetRequestJSON(payloadStruct interface{}, c echo.Context) error {
	body, err := GetRequestBody(c)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), payloadStruct)
	if err != nil {
		return err
	}

	return nil
}
