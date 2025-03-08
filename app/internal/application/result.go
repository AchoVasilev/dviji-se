package application

type Result[T any] struct {
	Success    bool
	Body       []T
	Message    string
	StatusCode int
	Error      error
}

type SingleResult[T any] struct {
	Success    bool
	Body       T
	Message    string
	StatusCode int
	Error      error
}

func ErrorSingleResult[T error](err T, message string, statusCode int) *SingleResult[T] {
	return &SingleResult[T]{
		Success:    false,
		Body:       err,
		Message:    message,
		StatusCode: statusCode,
		Error:      err,
	}
}
