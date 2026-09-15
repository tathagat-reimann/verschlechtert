package util

import "net/http"

type ErrorCode string

const (
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrInternalError  ErrorCode = "INTERNAL_ERROR"

	MissingBearerToken    ErrorCode = "MISSING_BEARER_TOKEN"
	ErrLoadingUser        ErrorCode = "LOADING_USER"
	InvalidLocale         ErrorCode = "INVALID_LOCALE"
	ErrUpdateLocale       ErrorCode = "UPDATE_LOCALE"
	Unauthorized          ErrorCode = "UNAUTHORIZED"
	ErrLoadingAllReports  ErrorCode = "LOADING_ALL_REPORTS"
	ReportNotFound        ErrorCode = "REPORT_NOT_FOUND"
	ErrJsonDecoding       ErrorCode = "JSON_DECODING"
	ErrInvalidReportImage ErrorCode = "INVALID_REPORT_IMAGE"
	ErrCreateReport       ErrorCode = "CREATE_REPORT"
	ErrAddComment         ErrorCode = "ADD_COMMENT"
	ErrToggleLike         ErrorCode = "TOGGLE_LIKE"
	ErrInvalidAlternative ErrorCode = "INVALID_ALTERNATIVE"
	ErrAddAlternative     ErrorCode = "ADD_ALTERNATIVE"
	ErrInvalidReportID    ErrorCode = "INVALID_REPORT_ID"
)

type ErrorDefinition struct {
	Code    ErrorCode
	Message string
	Status  int
}

var ErrorMessages = map[ErrorCode]ErrorDefinition{
	ErrInvalidRequest: {
		Code:    ErrInvalidRequest,
		Message: "Invalid request",
		Status:  400,
	},
	ErrNotFound: {
		Code:    ErrNotFound,
		Message: "Resource not found",
		Status:  404,
	},
	ErrInternalError: {
		Code:    ErrInternalError,
		Message: "Internal server error",
		Status:  500,
	},
	MissingBearerToken: {
		Code:    MissingBearerToken,
		Message: "Missing bearer token",
		Status:  http.StatusUnauthorized,
	},
	ErrLoadingUser: {
		Code:    ErrLoadingUser,
		Message: "could not load user",
		Status:  http.StatusInternalServerError,
	},
	InvalidLocale: {
		Code:    InvalidLocale,
		Message: "invalid locale",
		Status:  http.StatusBadRequest,
	},
	ErrUpdateLocale: {
		Code:    ErrUpdateLocale,
		Message: "could not update user locale",
		Status:  http.StatusInternalServerError,
	},
	Unauthorized: {
		Code:    Unauthorized,
		Message: "unauthorized",
		Status:  http.StatusUnauthorized,
	},
	ErrLoadingAllReports: {
		Code:    ErrLoadingAllReports,
		Message: "could not load all reports",
		Status:  http.StatusInternalServerError,
	},
	ReportNotFound: {
		Code:    ReportNotFound,
		Message: "report not found",
		Status:  http.StatusNotFound,
	},
	ErrJsonDecoding: {
		Code:    ErrJsonDecoding,
		Message: "JSON decoding error",
		Status:  http.StatusBadRequest,
	},
	ErrInvalidReportImage: {
		Code:    ErrInvalidReportImage,
		Message: "invalid report image",
		Status:  http.StatusBadRequest,
	},
	ErrCreateReport: {
		Code:    ErrCreateReport,
		Message: "could not create report",
		Status:  http.StatusInternalServerError,
	},
	ErrAddComment: {
		Code:    ErrAddComment,
		Message: "could not add comment",
		Status:  http.StatusInternalServerError,
	},
	ErrToggleLike: {
		Code:    ErrToggleLike,
		Message: "could not toggle like",
		Status:  http.StatusInternalServerError,
	},
	ErrInvalidAlternative: {
		Code:    ErrInvalidAlternative,
		Message: "invalid alternative",
		Status:  http.StatusBadRequest,
	},
	ErrAddAlternative: {
		Code:    ErrAddAlternative,
		Message: "could not add alternative",
		Status:  http.StatusInternalServerError,
	},
	ErrInvalidReportID: {
		Code:    ErrInvalidReportID,
		Message: "invalid report id",
		Status:  http.StatusBadRequest,
	},
}
