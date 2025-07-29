package errors

import (
	"net/http"

	"github.com/joomcode/errorx"
)

type ErrorType struct {
	StatusCode int
	Type       *errorx.Type
}

var Error = []ErrorType{
	{
		StatusCode: http.StatusBadRequest,
		Type:       ErrInvalidUserInput,
	},
	{
		StatusCode: http.StatusForbidden,
		Type:       ErrAcessError,
	},
	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrInternalServerError,
	},
	{
		StatusCode: http.StatusBadRequest,
		Type:       ErrAuthClient,
	},

	{
		StatusCode: http.StatusUnauthorized,
		Type:       ErrInvalidAccessToken,
	},
	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrUnableToGet,
	},

	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrUnableTocreate,
	},
	{
		StatusCode: http.StatusNotFound,
		Type:       ErrResourceNotFound,
	},

	{
		StatusCode: http.StatusConflict,
		Type:       ErrDataAlredyExist,
	},
	{
		StatusCode: http.StatusConflict,
		Type:       ErrDuplicatePaymentReference,
	},
	{
		StatusCode: http.StatusNotFound,
		Type:       ErrNoRecordFound,
	},
	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrDBDelError,
	},

	{
		StatusCode: http.StatusGone,
		Type:       ErrSocketConnectionReset,
	},
	{
		StatusCode: http.StatusGone,
		Type:       ErrSocketConnectionBroken,
	},
	{
		StatusCode: http.StatusRequestTimeout,
		Type:       ErrDeadlineTimedOut,
	},
	{
		StatusCode: http.StatusBadRequest,
		Type:       ErrProgramStatus,
	},
	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrPISISMSError,
	},
	{
		StatusCode: http.StatusInternalServerError,
		Type:       ErrRedisClientError,
	},
}

// list of error namespaces
var (
	databaseError           = errorx.NewNamespace("database error").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	invalidInput            = errorx.NewNamespace("validation error").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	resourceNotFound        = errorx.NewNamespace("not found").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	unauthorized            = errorx.NewNamespace("unauthorized").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	ineligible              = errorx.NewNamespace("ineligible").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	AccessDenied            = errorx.RegisterTrait("You are not authorized to perform the action")
	Ineligible              = errorx.RegisterTrait("You are not eligible to perform the action")
	serverError             = errorx.NewNamespace("server error")
	authoriztionClientError = errorx.NewNamespace("authorization client error")
	Unauthenticated         = errorx.NewNamespace("user authentication failed")
	ProgramError            = errorx.NewNamespace("program error")
	websocketError          = errorx.NewNamespace("websocket error")
	httpError               = errorx.NewNamespace("http error")
	dbError                 = errorx.NewNamespace("db error")
	duplicate               = errorx.NewNamespace("duplicate").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	externalserviceFailed   = errorx.NewNamespace("external service Failed").
				ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	logicFailed    = errorx.NewNamespace("business logic Failed").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	requestFailed  = errorx.NewNamespace("request binding Failed").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	bodyreadFailed = errorx.NewNamespace("reading response body Failed").
			ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	unroutableLocation = errorx.NewNamespace("unroutable location").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	kafkaError         = errorx.NewNamespace("kafka error").ApplyModifiers(errorx.TypeModifierOmitStackTrace)
)

// list of errors types in all of the above namespaces

var (
	ErrUnableTocreate          = errorx.NewType(databaseError, "unable to create")
	ErrDataAlredyExist         = errorx.NewType(databaseError, "data alredy exist")
	ErrUnableToGet             = errorx.NewType(databaseError, "unable to get")
	ErrInvalidUserInput        = errorx.NewType(invalidInput, "invalid user input")
	ErrInactiveUserStatus      = errorx.NewType(invalidInput, "Inactive user status")
	ErrTripDeviceChange        = errorx.NewType(invalidInput, "user changed device")
	ErrResourceNotFound        = errorx.NewType(resourceNotFound, "resource not found")
	ErrAcessError              = errorx.NewType(unauthorized, "Unauthorized", AccessDenied)
	ErrIneligibleError         = errorx.NewType(ineligible, "Ineligible", Ineligible)
	ErrInternalServerError     = errorx.NewType(serverError, "internal server error")
	ErrAuthClient              = errorx.NewType(authoriztionClientError, "authorization client error")
	ErrSSOAuthenticationFailed = errorx.NewType(Unauthenticated, "user authentication failed")
	ErrInvalidAccessToken      = errorx.NewType(Unauthenticated, "invalid token").
					ApplyModifiers(errorx.TypeModifierOmitStackTrace)
	ErrSSOError                  = errorx.NewType(serverError, "sso communication failed")
	ErrAccountingError           = errorx.NewType(serverError, "accounting error")
	ErrUnExpectedError           = errorx.NewType(serverError, "unexpected error occurred")
	ErrUnableToUpdate            = errorx.NewType(databaseError, "unable to update")
	ErrDBDelError                = errorx.NewType(databaseError, "could not delete record")
	ErrNoRecordFound             = errorx.NewType(resourceNotFound, "no record found")
	ErrEventNotSupported         = errorx.NewType(websocketError, "event type not supported")
	ErrDeadlineTimedOut          = errorx.NewType(websocketError, "read or write deadline timedout")
	ErrSocketConnectionClosed    = errorx.NewType(websocketError, "socket client connection closed")
	ErrSocketConnectionBroken    = errorx.NewType(websocketError, "broken pipe")
	ErrSocketConnectionReset     = errorx.NewType(websocketError, "connection reset by client")
	ErrProgramStatus             = errorx.NewType(ProgramError, "program status error")
	ErrProgramAmount             = errorx.NewType(ProgramError, "spending limit error")
	ErrSMSSend                   = errorx.NewType(serverError, "couldn't send sms")
	ErrHTTPRequestPrepareFailed  = errorx.NewType(httpError, "couldn't prepare http request")
	ErrSocketReadLimitExceeded   = errorx.NewType(websocketError, "socket read limit exceeded")
	ErrWriteError                = errorx.NewType(dbError, "could not write to db")
	ErrReadError                 = errorx.NewType(dbError, "could not read data from db")
	ErrDataExists                = errorx.NewType(duplicate, "data already exists")
	ErrDuplicatePaymentReference = errorx.NewType(duplicate, "payment reference already processed")
	ErrExternalServiceFailed     = errorx.NewType(externalserviceFailed, "external service has Failed")
	ErrLogicFailed               = errorx.NewType(logicFailed, "logic failure")
	ErrHTTPRequestBinding        = errorx.NewType(requestFailed, "binding failure")
	ErrReadingResponseBody       = errorx.NewType(bodyreadFailed, "reading body failure")
	ErrUnroutableLocaiton        = errorx.NewType(unroutableLocation, "Failed to find route for location")
	ErrSendNotificationFailed    = errorx.NewType(httpError, "unable to  send http request")
	UnexpectedError              = errorx.NewType(serverError, "invalid value")
	ErrKafkaEventNotSupported    = errorx.NewType(kafkaError, "event type not supported")
	ErrPISISMSError              = errorx.NewType(serverError, "unable to send sms message to the user")
	ErrRedisClientError          = errorx.NewType(serverError, "redis client error")
)
