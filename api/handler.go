package api

import (
	"fmt"
	"net/http"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprint(w, "success")
	return nil
}
