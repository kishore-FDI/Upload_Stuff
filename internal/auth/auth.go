package auth

import (
	"encoding/json"
	"net/http"
)

type authRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

func SignUp(w http.ResponseWriter,r *http.Request){
	if r.Method!=http.MethodPost{
		http.Error(w,"Method Not Allowed. Only Post Method is Allowed",http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err:= decoder.Decode(&req);err!=nil{
		http.Error(w,"Invalid Request: "+err.Error(),http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
}

func SignIn(){

}