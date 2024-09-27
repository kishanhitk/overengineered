package models

type NameRequest struct {
	Name string `json:"name"`
}

type GreetingResponse struct {
	Message string `json:"message"`
}
