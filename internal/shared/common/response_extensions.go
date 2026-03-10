// internal/shared/common/response_extensions.go
package common

import (
	"github.com/farmanexo/order-service/internal/shared/constants"
)

// ResponseBuilder proporciona métodos fluent para construir responses
type ResponseBuilder[T any] struct {
	response *ApiResponse[T]
}

func NewResponseBuilder[T any]() *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		response: NewApiResponse[T](),
	}
}

func (b *ResponseBuilder[T]) WithError(
	code constants.MessageCode,
	message string,
	statusCode constants.HTTPStatusCode,
) *ResponseBuilder[T] {
	b.response.SetHttpStatus(statusCode.Int())
	b.response.AddError(code, message)
	return b
}

func (b *ResponseBuilder[T]) WithSuccess(
	data T,
	statusCode constants.HTTPStatusCode,
) *ResponseBuilder[T] {
	b.response.SetData(data)
	b.response.AddSuccessMessage()
	b.response.SetHttpStatus(statusCode.Int())
	return b
}

func (b *ResponseBuilder[T]) WithData(data T) *ResponseBuilder[T] {
	b.response.SetData(data)
	return b
}

func (b *ResponseBuilder[T]) WithMessage(
	code constants.MessageCode,
	message string,
	messageType constants.MessageType,
) *ResponseBuilder[T] {
	b.response.AddMessageWithType(code, message, messageType)
	return b
}

func (b *ResponseBuilder[T]) WithHttpStatus(statusCode constants.HTTPStatusCode) *ResponseBuilder[T] {
	b.response.SetHttpStatus(statusCode.Int())
	return b
}

func (b *ResponseBuilder[T]) Build() *ApiResponse[T] {
	return b.response
}

func OkResponse[T any](data T) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithSuccess(data, constants.StatusOK).
		Build()
}

func CreatedResponse[T any](data T) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusCreated.Int())
	resp.SetData(data)
	resp.AddMessageWithType(
		constants.CodeCreatedSuccess,
		constants.GetDescription(constants.CodeCreatedSuccess),
		constants.MessageTypeSuccess,
	)
	return resp
}

func NoContentResponse[T any]() *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusNoContent.Int())
	return resp
}

func BadRequestResponse[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusBadRequest.Int())
	resp.AddError(code, message)
	return resp
}

func UnauthorizedResponse[T any](message string) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithError(constants.CodeUnauthorized, message, constants.StatusUnauthorized).
		Build()
}

func ForbiddenResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusForbidden.Int())
	resp.AddError(constants.CodeForbidden, message)
	return resp
}

func NotFoundResponse[T any](message string) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithError(constants.CodeResourceNotFound, message, constants.StatusNotFound).
		Build()
}

func ConflictResponse[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusConflict.Int())
	resp.AddError(code, message)
	return resp
}

func TooManyRequestsResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusTooManyRequests.Int())
	resp.AddError(constants.CodeRateLimitExceeded, message)
	return resp
}

func InternalServerErrorResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusInternalServerError.Int())
	resp.AddError(constants.CodeInternalError, message)
	return resp
}
