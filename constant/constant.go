package constant

type VerificationCodeGenerationMsg string

const (
	MsgDone                 VerificationCodeGenerationMsg = "done"
	MsgInvalidArguments     VerificationCodeGenerationMsg = "invalid argument(s)"
	MsgSendingTooFrequently VerificationCodeGenerationMsg = "sending too frequently"
	MsgNeedingResending     VerificationCodeGenerationMsg = "needing resending"
)

type VerificationCodeValidationMsg string

const (
	MsgValid           VerificationCodeValidationMsg = "valid"
	MsgInvalid         VerificationCodeValidationMsg = "invalid"
	MsgExpired         VerificationCodeValidationMsg = "expired"
	MsgMaximumAttempts VerificationCodeValidationMsg = "maximum attempts"
)
