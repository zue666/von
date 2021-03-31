package von

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/julienschmidt/httprouter"
)

var validate = validator.New()

var translator *ut.UniversalTranslator

func init() {
	enLocale := en.New()
	translator = ut.New(enLocale, enLocale)
	lang, _ := translator.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, lang)

	validate.RegisterValidation("notblank", validators.NotBlank)
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

}

// Params holds route params
type Params struct {
	httprouter.Params
}

// ByName returns the value of the given param key
func (params *Params) ByName(name string) string {
	return params.Params.ByName(name)
}

// ParamsFromContext returns params slice from request context
func ParamsFromContext(ctx context.Context) *Params {
	v, _ := ctx.Value(ParamsKey).(*Params)
	return v
}

// Decode reads the body of an HTTP request looking for a JSON document.
// The body is decoded into the provided value.
// If the provided value is a struct then it is checked for validation tags.
func Decode(r *http.Request, val interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		return NewRequestError(err, http.StatusBadRequest)
	}

	if err := validate.Struct(val); err != nil {
		verrors, ok := err.(validator.ValidationErrors)

		if !ok {
			return err
		}

		lang, _ := translator.GetTranslator("en")

		var fields []FieldError
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(lang),
			}
			fields = append(fields, field)
		}

		return &Error{
			Err:    errors.New("field validation error"),
			Status: http.StatusBadRequest,
			Fields: fields,
		}
	}
	return nil
}
