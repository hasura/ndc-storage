package functions

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-storage/connector/functions/internal"
)

var GetBaseConnectorSchema = internal.GetConnectorSchema

var errPermissionDenied = schema.ForbiddenError("permission dennied", nil)
