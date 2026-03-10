// pkg/mediator/mediator.go
package mediator

import (
	"context"
	"fmt"

	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
)

// Request es la interfaz base para Commands y Queries
type Request[TResponse any] interface {
	GetName() string
}

// RequestHandler maneja un Request y retorna una Response
type RequestHandler[TRequest Request[TResponse], TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (*common.ApiResponse[TResponse], error)
}

// Validator valida un Request antes de ejecutarlo
type Validator[TRequest Request[TResponse], TResponse any] interface {
	Validate(ctx context.Context, request TRequest) error
}

// PreProcessor ejecuta lógica antes del Handler
type PreProcessor[TRequest Request[TResponse], TResponse any] interface {
	Process(ctx context.Context, request TRequest) error
}

// PostProcessor ejecuta lógica después del Handler
type PostProcessor[TRequest Request[TResponse], TResponse any] interface {
	Process(ctx context.Context, request TRequest, response *common.ApiResponse[TResponse]) error
}

// Mediator es el coordinador central de CQRS
type Mediator struct {
	handlers       map[string]interface{}
	validators     map[string]interface{}
	preProcessors  []interface{}
	postProcessors []interface{}
}

// NewMediator crea una nueva instancia del Mediator
func NewMediator() *Mediator {
	return &Mediator{
		handlers:       make(map[string]interface{}),
		validators:     make(map[string]interface{}),
		preProcessors:  make([]interface{}, 0),
		postProcessors: make([]interface{}, 0),
	}
}

// RegisterHandler registra un handler para un request específico
func RegisterHandler[TRequest Request[TResponse], TResponse any](
	m *Mediator,
	handler RequestHandler[TRequest, TResponse],
) {
	var req TRequest
	requestName := getTypeName(req)
	m.handlers[requestName] = handler
}

// RegisterValidator registra un validator
func RegisterValidator[TRequest Request[TResponse], TResponse any](
	m *Mediator,
	validator Validator[TRequest, TResponse],
) {
	var req TRequest
	requestName := getTypeName(req)
	m.validators[requestName] = validator
}

// RegisterPreProcessor registra un preprocessor global
func (m *Mediator) RegisterPreProcessor(processor interface{}) {
	m.preProcessors = append(m.preProcessors, processor)
}

// RegisterPostProcessor registra un postprocessor global
func (m *Mediator) RegisterPostProcessor(processor interface{}) {
	m.postProcessors = append(m.postProcessors, processor)
}

// Send ejecuta el pipeline completo: Validation -> PreProcessors -> Handler -> PostProcessors
func Send[TRequest Request[TResponse], TResponse any](
	ctx context.Context,
	m *Mediator,
	request TRequest,
) (*common.ApiResponse[TResponse], error) {
	requestName := getTypeName(request)

	// 1. VALIDATION
	if err := m.runValidation(ctx, requestName, request); err != nil {
		return createValidationErrorResponse[TResponse](err), nil
	}

	// 2. PRE-PROCESSORS
	if err := m.runPreProcessors(ctx, request); err != nil {
		return createErrorResponse[TResponse](
			constants.CodeInternalError,
			"Error en pre-procesamiento: "+err.Error(),
		), nil
	}

	// 3. HANDLER
	handler, exists := m.handlers[requestName]
	if !exists {
		return createErrorResponse[TResponse](
			constants.CodeInternalError,
			fmt.Sprintf("No handler registrado para: %s", requestName),
		), fmt.Errorf("handler not found for %s", requestName)
	}

	typedHandler, ok := handler.(RequestHandler[TRequest, TResponse])
	if !ok {
		return createErrorResponse[TResponse](
			constants.CodeInternalError,
			"Error de tipo en handler",
		), fmt.Errorf("handler type mismatch")
	}

	response, err := typedHandler.Handle(ctx, request)
	if err != nil {
		return createErrorResponse[TResponse](
			constants.CodeInternalError,
			err.Error(),
		), err
	}

	// 4. POST-PROCESSORS
	if err := m.runPostProcessors(ctx, request, response); err != nil {
		fmt.Printf("Error en post-procesamiento: %v\n", err)
	}

	return response, nil
}

func (m *Mediator) runValidation(ctx context.Context, requestName string, request interface{}) error {
	if validator, exists := m.validators[requestName]; exists {
		if v, ok := validator.(interface {
			Validate(context.Context, interface{}) error
		}); ok {
			return v.Validate(ctx, request)
		}
	}
	return nil
}

func (m *Mediator) runPreProcessors(ctx context.Context, request interface{}) error {
	for _, processor := range m.preProcessors {
		if p, ok := processor.(interface {
			Process(context.Context, interface{}) error
		}); ok {
			if err := p.Process(ctx, request); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Mediator) runPostProcessors(ctx context.Context, request interface{}, response interface{}) error {
	for _, processor := range m.postProcessors {
		if p, ok := processor.(interface {
			Process(context.Context, interface{}, interface{}) error
		}); ok {
			if err := p.Process(ctx, request, response); err != nil {
				return err
			}
		}
	}
	return nil
}

func getTypeName(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func createValidationErrorResponse[TResponse any](err error) *common.ApiResponse[TResponse] {
	resp := common.NewApiResponse[TResponse]()
	resp.SetHttpStatus(constants.StatusBadRequest.Int())
	resp.AddError(constants.CodeValidationError, err.Error())
	return resp
}

func createErrorResponse[TResponse any](code constants.MessageCode, message string) *common.ApiResponse[TResponse] {
	resp := common.NewApiResponse[TResponse]()
	resp.SetHttpStatus(constants.StatusInternalServerError.Int())
	resp.AddError(code, message)
	return resp
}

// Pipeline Context
type pipelineContextKey string

const (
	UserIDKey      pipelineContextKey = "user_id"
	CorrelationKey pipelineContextKey = "correlation_id"
	RequestNameKey pipelineContextKey = "request_name"
)

func WithValue(ctx context.Context, key pipelineContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetValue(ctx context.Context, key pipelineContextKey) interface{} {
	return ctx.Value(key)
}

func GetUserID(ctx context.Context) (string, bool) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}

func GetCorrelationID(ctx context.Context) string {
	val := ctx.Value(CorrelationKey)
	if val == nil {
		return ""
	}
	corrID, _ := val.(string)
	return corrID
}
