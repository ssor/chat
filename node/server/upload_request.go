package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, fileName string, file io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	part, err := writer.CreateFormFile(paramName, fileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	if params != nil {
		for key, val := range params {
			_ = writer.WriteField(key, val)
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	return req, nil
}

func uploadFile(url, fileName string, extraParams map[string]string, file io.Reader) (*http.Response, []byte, error) {
	request, err := newfileUploadRequest(url, extraParams, "file", fileName, file)
	if err != nil {
		return nil, nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	// Reset resp.Body so it can be use again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// deep copy response to give it to both return and callback func
	return resp, body, nil
}
