package scanner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

//Tasker was tasker list of sqlmap
type Tasker struct {
	Status  string `json:"status"`
	Taskid  string `json:"taskid"`
	URL     string `json:"url"`
	Success bool   `json:"success"`
	// Returncode int      `json:"returncode"`
	Data  []string `json:"data"`
	Error []string `json:"error"`
}

//LocalServer set the server
const LocalServer = "http://127.0.0.1:8775/"

const scanURL = "https://www.douban.com/accounts/login"

var c *http.Client

//SetTaskID set the Tasker's Taskerid
func (Task *Tasker) SetTaskID(tid string) {
	Task.Taskid = tid
}

//NewTasker call a new Tasker  if success it will return a taskid ,otherwise, an error.
func NewTasker() (string, error) {
	var jstring Tasker
	c = &http.Client{}
	req, err := http.NewRequest("GET", LocalServer+"task/new", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &jstring)
	if err != nil {
		return "", err
	}

	return jstring.Taskid, nil
}

//DeleteTasker delete a task with TID
func DeleteTasker(taskerid string) (bool, error) {
	var jstring map[string]bool
	c = &http.Client{}
	req, err := http.NewRequest("GET", LocalServer+"task/"+taskerid+"/delete", nil)
	if err != nil {
		return false, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(body, &jstring)
	if err != nil {
		return false, err
	}

	return jstring["success"], nil
}

//StartTasker config the sqlmap process
func StartTasker(taskerid string, scanURL string) (bool, error) {
	var result Tasker
	data := map[string]string{
		"url": scanURL,
	}
	datajson, err := json.Marshal(data)
	if err != nil {
		return false, nil
	}
	req, err := http.NewRequest("POST", LocalServer+"scan/"+taskerid+"/start", strings.NewReader(string(datajson)))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return false, nil
	}
	resp, err := c.Do(req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, nil
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, err
	}

	return result.Success, nil
}

//GetTaskerStatus hook the current status of running tasker
func GetTaskerStatus(taskerid string) (string, error) {
	// var result Tasker
	req, err := http.NewRequest("GET", LocalServer+"scan/"+taskerid+"/status", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

//GetResultFromTasker   finally fetch the result
func GetResultFromTasker(taskerid string) ([]string, error) {
	var result Tasker
	req, err := http.NewRequest("GET", LocalServer+"scan/"+taskerid+"/data", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
