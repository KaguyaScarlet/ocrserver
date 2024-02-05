package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"github.com/otiai10/marmoset"
)

// Base64 ...
func Base64(w http.ResponseWriter, r *http.Request) {

	render := marmoset.Render(w, true)

	var body = new(struct {
		Base64    string `json:"base64"`
		Trim      string `json:"trim"`
		Languages string `json:"languages"`
		Whitelist string `json:"whitelist"`
		PSM       string `json:"psm"`
	})

	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		render.JSON(http.StatusBadRequest, err)
		return
	}

	tempfile, err := ioutil.TempFile("", "ocrserver"+"-")
	if err != nil {
		render.JSON(http.StatusInternalServerError, err)
		return
	}
	defer func() {
		tempfile.Close()
		os.Remove(tempfile.Name())
	}()

	if len(body.Base64) == 0 {
		render.JSON(http.StatusBadRequest, fmt.Errorf("base64 string required"))
		return
	}
	body.Base64 = regexp.MustCompile("data:image\\/png;base64,").ReplaceAllString(body.Base64, "")
	b, err := base64.StdEncoding.DecodeString(body.Base64)
	if err != nil {
		render.JSON(http.StatusBadRequest, err)
		return
	}
	tempfile.Write(b)

	client := gosseract.NewClient()
	defer client.Close()

	client.Languages = []string{"eng"}
	if body.Languages != "" {
		client.Languages = strings.Split(body.Languages, ",")
	}
	client.SetImage(tempfile.Name())
	if body.Whitelist != "" {
		client.SetWhitelist(body.Whitelist)
	}

	if body.PSM != "" {
		psm, err := strconv.Atoi(body.PSM)
		if err != nil {
			render.JSON(http.StatusBadRequest, "psm is not a number")
			return
		}

		if psm < 0 || psm > 13 {
			render.JSON(http.StatusBadRequest, "psm is out of range，please enter a number between 0 and 13")
			return
		}

		fmt.Printf("set psm %d\n", psm)
		client.SetVariable("tessedit_pageseg_mode", body.PSM) // client.SetPageSegMode(psm) not working
	}

	text, err := client.Text()
	if err != nil {
		render.JSON(http.StatusInternalServerError, err)
		return
	}

	render.JSON(http.StatusOK, map[string]interface{}{
		"result":  strings.Trim(text, body.Trim),
		"version": version,
	})
}
