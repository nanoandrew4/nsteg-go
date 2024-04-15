package server

import "nsteg/api"

var (
	errRequestBodyDecode       = api.Error{Error: "Error reading request body"}
	errInvalidImage            = api.Error{Code: "invalid_image", Error: "Invalid image supplied in request body"}
	errErrorConvertingRawImage = api.Error{Error: "Error converting raw image"}
)
