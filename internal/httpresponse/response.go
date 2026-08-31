package httpresponse

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}


func Error(message string) Response{
	return Response{
		Status: "error",
		Message: message,
	}
}

func Success(data interface{}, message string) Response{
	return Response{
		Status: "success",
		Message: message,
		Data: data,
	}
}