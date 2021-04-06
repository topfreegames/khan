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

	"github.com/spf13/viper"

	"github.com/labstack/echo"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
	"github.com/topfreegames/khan/log"
	"github.com/topfreegames/khan/models"
	"github.com/uber-go/zap"
)

//EasyJSONUnmarshaler describes a struct able to unmarshal json
type EasyJSONUnmarshaler interface {
	UnmarshalEasyJSON(l *jlexer.Lexer)
}

//EasyJSONMarshaler describes a struct able to marshal json
type EasyJSONMarshaler interface {
	MarshalEasyJSON(w *jwriter.Writer)
}

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	payload := map[string]interface{}{
		"success": false,
		"reason":  message,
	}
	return c.JSON(status, payload)
}

// FailWithError fails with the specified error
func FailWithError(err error, c echo.Context) error {
	t := reflect.TypeOf(err)
	status, ok := map[string]int{
		"*models.ModelNotFoundError":                                 http.StatusNotFound,
		"*models.PlayerReachedMaxInvitesError":                       http.StatusBadRequest,
		"*models.ForbiddenError":                                     http.StatusForbidden,
		"*models.PlayerCannotPerformMembershipActionError":           http.StatusForbidden,
		"*models.AlreadyHasValidMembershipError":                     http.StatusConflict,
		"*models.CannotApproveOrDenyMembershipAlreadyProcessedError": http.StatusConflict,
		"*models.CannotPromoteOrDemoteMemberLevelError":              http.StatusConflict,
	}[t.String()]

	if !ok {
		status = http.StatusInternalServerError
	}

	return FailWith(status, err.Error(), c)
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	f := func() error {
		payload["success"] = true
		return c.JSON(http.StatusOK, payload)
	}
	return f()
}

//LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c echo.Context, logger zap.Logger) error {
	log.D(logger, "Loading payload...")

	data, err := GetRequestBody(c)
	if err != nil {
		log.E(logger, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	unmarshaler, ok := payloadStruct.(EasyJSONUnmarshaler)
	if !ok {
		err := fmt.Errorf("Can't unmarshal specified payload since it does not implement easyjson interface")
		log.E(logger, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	lexer := jlexer.Lexer{Data: []byte(data)}
	unmarshaler.UnmarshalEasyJSON(&lexer)
	if err = lexer.Error(); err != nil {
		log.E(logger, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	if validatable, ok := payloadStruct.(Validatable); ok {
		missingFieldErrors := validatable.Validate()

		if len(missingFieldErrors) != 0 {
			err := errors.New(strings.Join(missingFieldErrors[:], ", "))
			log.E(logger, "Loading payload failed.", func(cm log.CM) {
				cm.Write(zap.Error(err))
			})
			return err
		}
	}

	log.D(logger, "Payload loaded successfully.")
	return nil
}

//GetRequestBody from echo context
func GetRequestBody(c echo.Context) ([]byte, error) {
	bodyCache := c.Get("requestBody")
	if bodyCache != nil {
		return bodyCache.([]byte), nil
	}
	body := c.Request().Body()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	c.Set("requestBody", b)
	return b, nil
}

//GetRequestJSON as the specified interface from echo context
func GetRequestJSON(payloadStruct interface{}, c echo.Context) error {
	body, err := GetRequestBody(c)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, payloadStruct)
	if err != nil {
		return err
	}

	return nil
}

// SetRetrieveClanHandlerConfigurationDefaults sets the default configs for RetrieveClanHandler
func SetRetrieveClanHandlerConfigurationDefaults(config *viper.Viper) {
	config.SetDefault(models.MaxPendingApplicationsKey, 100)
	config.SetDefault(models.MaxPendingInvitesKey, 100)
	config.SetDefault(models.PendingApplicationsOrderKey, models.Newest)
	config.SetDefault(models.PendingInvitesOrderKey, models.Newest)
}
