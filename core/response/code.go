package response

type ResponseCode string

const (
	CodeOK         ResponseCode = "OK"
	CodeCreated    ResponseCode = "CREATED"
	CodeUpdated    ResponseCode = "UPDATED"
	CodeNotFound   ResponseCode = "NOT_FOUND"
	CodeDuplicate  ResponseCode = "DUPLICATE"
	CodeBadRequest ResponseCode = "BAD_REQUEST"
	CodeInternal   ResponseCode = "INTERNAL_ERROR"
)
