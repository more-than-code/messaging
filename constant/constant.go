package constant

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	UserRoleCommon   = 0
	UserRoleRetailer = 1
	UserRoleEditor   = 2
	UserRoleAdmin    = 3
)

const (
	PurposeRegistration        = 0
	PurposeLogin               = 1
	PurposeResettingPassword   = 2
	PurposeBindingPhoneOrEmail = 3
)

const (
	ExHttpStatusAuthenticationFailure = 460
	ExHttpStatusBusinessError         = 461
)

const (
	CodeExistingUser    = 1000
	CodeNonexistingUser = 1001

	CodeAuthenticationFailure = 1002
	CodeUnauthorized          = 1003
	CodeNotAllowed            = 1004

	CodeExpiredVerificationCode              = 1100
	CodeWrongVerificationCode                = 1101
	CodeTooFrequentlySendingVerificationCode = 1102
	CodeNeedingResendingVerificationCode     = 1103
	CodeMaximumAttemptsOnVerificationCode    = 1104

	CodeInvalidParameters    = 1200
	CodeWrongEmailOrPassword = 1201
	CodeDeletedUser          = 1202

	CodeDatabaseOperationFailure = 2000
	CodeDatabaseDocumentNotFound = 2001
	CodeDataUnmarshalingFailure  = 2002
	CodeDataMarshalingFailure    = 2003

	CodeUnderlyingRequestFailure = 3000

	CodeServiceInternalServerError = 5000
	CodeServiceNotImplemented      = 5001
	CodeServiceBadGateway          = 5002
	CodeServiceServiceUnavailable  = 5003
	CodeServiceGatewayTimeout      = 5004
)

const (
	MsgExistingUser          = "user already exists"
	MsgNonexistingUser       = "user not exists"
	MsgAuthenticationFailure = "authenticating user failure"
	MsgUnauthorized          = "unauthorized "
	MsgNotAllowed            = "not allowed"

	MsgExpiredVerificationCode              = "expired verification code"
	MsgWrongVerificationCode                = "wrong verification code"
	MsgDeletedUser                          = "deleted user"
	MsgTooFrequentlySendingVerificationCode = "too frequently sending code"
	MsgNeedingResendingVerificationCode     = "needing resending verification code"
	MsgMaximumAttemptsOnVerificationCode    = "maximum attempts on verification code"

	MsgInvalidParamters     = "invalid parameter(s)"
	MsgWrongEmailOrPassword = "wrong email or password"

	MsgDatabaseOperationFailure = "database operation failure"
	MsgDatabaseDocumentNotFound = "document not found in database"
	MsgDataUnmarshalingFailure  = "data unmarshaling failure"
	MsgDataMarshalingFailure    = "data marshaling failure"

	MsgUnderlyingRequestFailure = "target request failure"
)

func TranslateToHttpStatusCode(code int) int {
	switch code {
	case int(codes.Unavailable):
		return http.StatusServiceUnavailable
	case CodeDatabaseDocumentNotFound:
		return http.StatusNotFound
	}

	if code < 5000 {
		return ExHttpStatusBusinessError
	}

	return http.StatusInternalServerError
}

func TranslateToServiceCode(code int) int {
	switch code {
	case int(codes.Unavailable):
		return CodeServiceServiceUnavailable
	default:
		return code
	}

}
