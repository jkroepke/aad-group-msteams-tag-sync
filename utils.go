package main

import (
	"errors"
	"fmt"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
)

func GetOdataError(err error) error {
	if err == nil {
		return err
	}

	var typed *odataerrors.ODataError
	if errors.As(err, &typed) {
		if terr := typed.GetErrorEscaped(); terr != nil {
			return fmt.Errorf("%w - %s: %s", err, *terr.GetCode(), *terr.GetMessage())
		}
	}

	return err
}
