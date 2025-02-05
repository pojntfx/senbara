package controllers

import (
	"log"
	"net/http"
)

func (c *Controller) HandleCode(w http.ResponseWriter, r *http.Request, code []byte) {
	c.log.Debug("Getting embedded source code")

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="code.tar.gz"`)

	if _, err := w.Write(code); err != nil {
		log.Println(errCouldNotWriteResponse, err)

		http.Error(w, errCouldNotWriteResponse.Error(), http.StatusInternalServerError)

		return
	}
}
