package linker_util

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
    "net/url"
    "os"
)

func Http_get_resp(url string, contenttype string, token string, getbody string) (resp *http.Response, err error) {
	client := &http.Client{}

	var Body io.Reader

	if getbody == "" {
		Body = nil
	} else {
		Body = ioutil.NopCloser(strings.NewReader(getbody))
	}
	req, _ := http.NewRequest("GET", url, Body)
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("X-Auth-Token", token)
	resp, err = client.Do(req)
	return
}

func Http_get(url string, contenttype string, token string, getbody string) (response string, err error) {
	client := &http.Client{}

	var Body io.Reader

	if getbody == "" {
		Body = nil
	} else {
		Body = ioutil.NopCloser(strings.NewReader(getbody))
	}
	req, _ := http.NewRequest("GET", url, Body)
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func Http_post(url string, contenttype string, token string, postbody string) (response string, err error) {
	client := &http.Client{}
	body := ioutil.NopCloser(strings.NewReader(postbody))
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func Http_put(url string, contenttype string, token string, postbody string) (response string, err error) {
	client := &http.Client{}
	body := ioutil.NopCloser(strings.NewReader(postbody))
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func Http_delete(url string, contenttype string, token string, postbody string) (response string, err error) {
	client := &http.Client{}
	body := ioutil.NopCloser(strings.NewReader(postbody))
	req, _ := http.NewRequest("DELETE", url, body)
	req.Header.Set("Content-Type", contenttype)
	req.Header.Set("X-Auth-Token", token)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func DownloadFile(fileurl string, targetFolder string, filename string) (err error) {
	fileURL, err := url.Parse(fileurl)
    if err != nil {
       return
    }
	var fileName string
	
	if filename == "" {
		path := fileURL.Path
    	segments := strings.Split(path, "/")
    	fileName = segments[2]
	} else {
		fileName = filename
	}
	
    file, err := os.Create(targetFolder + "/" + fileName)	
	check := http.Client{
        CheckRedirect: func(r *http.Request, via []*http.Request) error {
                r.URL.Opaque = r.URL.Path
                return nil
        },
    }

    resp, err := check.Get(fileurl) // add a filter to check redirect

    if err != nil {
        return
    }
    defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	
	return
}
