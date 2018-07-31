package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	perm = os.FileMode(0600)
	dir  = "api"
)

func main() {
	fname := flag.String("c", "xxx.json", "path to a Postman Collection")
	flag.Parse()

	data, err := ioutil.ReadFile(*fname)
	CheckErr(err)

	ps := new(Collection)
	CheckErr(json.Unmarshal(data, ps))

	CheckErr(os.MkdirAll(dir, 0755))

	GenerateMD(ps.Item, "", 1)
}

func CheckErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func GenerateMD(items []json.RawMessage, pName string, level int) {
	if len(items) == 0 {
		return
	}

	buf := bytes.NewBuffer(nil)

	if level == 1 {
		buf.WriteString(fmt.Sprintf("# %s", "api"))
	} else {
		buf.WriteString(fmt.Sprintf("# %s", pName))
	}

	var hasAPI bool

	for _, v := range items {
		tmp := new(APIItem)
		CheckErr(json.Unmarshal(v, tmp))

		isFolder := tmp.Request.Method == ""
		if isFolder {
			ftmp := new(FolderItem)
			CheckErr(json.Unmarshal(v, ftmp))

			if level == 1 {
				GenerateMD(ftmp.Item, tmp.Name, level+1)
			} else {
				GenerateMD(ftmp.Item, pName+"_"+tmp.Name, level+1)
			}
		} else {
			buf.Write(BuildAPI(tmp))

			hasAPI = true
		}
	}

	if hasAPI {
		if level == 1 {
			CheckErr(ioutil.WriteFile(filepath.Join(dir, "api.md"), buf.Bytes(), perm))
		} else {
			CheckErr(ioutil.WriteFile(filepath.Join(dir, pName+".md"), buf.Bytes(), perm))
		}
	}
}

func BuildAPI(item *APIItem) []byte {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("\n\n")

	title, remark := splitDescription(item.Request.Description)
	buf.WriteString(fmt.Sprintf("## %s\n", title))
	buf.WriteString(remark + "\n\n")

	buf.WriteString("```\n")
	buf.WriteString(fmt.Sprintf("%s %s\n", item.Request.Method, item.Request.URL.Raw))
	buf.WriteString("```\n")

	if len(item.Request.URL.Query) > 0 {
		buf.WriteString("\nreq query说明:\n")
		buf.WriteString("\n|key|value|说明|\n")
		buf.WriteString("|---|---|---|\n")
		for _, v := range item.Request.URL.Query {
			buf.WriteString(fmt.Sprintf("|%s|%s|%s|\n", v.Key, v.Value, v.Description))
		}
	}

	if len(item.Request.Header) > 0 {
		buf.WriteString("\nreq header说明:\n")
		buf.WriteString("\n|key|value|说明|\n")
		buf.WriteString("|---|---|---|\n")
		for _, v := range item.Request.Header {
			if v.Key == "X-MX-Token" {
				buf.WriteString(fmt.Sprintf("|%s|%s|%s|\n", v.Key, "token", "需要"))
			} else {
				buf.WriteString(fmt.Sprintf("|%s|%s|%s|\n", v.Key, v.Value, v.Description))
			}
		}
	}

	if item.Request.Method != "GET" && item.Request.Body.Raw != "" {
		buf.WriteString("\nreq body:\n")
		buf.WriteString("```json\n")
		buf.WriteString(fmt.Sprintf("%s\n", item.Request.Body.Raw))
		buf.WriteString("```\n")
	}

	return buf.Bytes()
}

func splitDescription(raw string) (string, string) {
	raw = strings.TrimSpace(raw)

	tmp := []rune(raw)
	splitChar := rune('\n')

	index := -1
	for i, v := range tmp {
		if v == splitChar {
			index = i
			break
		}
	}

	if index == -1 {
		return "unknow", raw
	}

	remark := strings.TrimSpace(string(tmp[index:]))

	return string(tmp[:index]), remark
}

// -- model
// Postman Collection v2.1
type Collection struct {
	Info CollectionInfo    `json:"info"`
	Item []json.RawMessage `json:"item"`
}

type CollectionInfo struct {
	Name        string `json:"name,omitempty"`
	PostmanID   string `json:"_postman_id,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema,omitempty"`
}

type FolderItem struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Item        []json.RawMessage `json:"item,omitempty"`
}

type APIItem struct {
	Name    string  `json:"name,omitempty"`
	Request Request `json:"request,omitempty"`
}

type Request struct {
	Method      string      `json:"method,omitempty"`
	Header      []Header    `json:"header,omitempty"`
	Body        RequestBody `json:"body,omitempty"`
	URL         URL         `json:"url,omitempty"`
	Description string      `json:"description,omitempty"`
}

type Header struct {
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

type RequestBody struct {
	Mode string `json:"mode,omitempty"`
	Raw  string `json:"raw,omitempty"`
}

type URL struct {
	Raw   string     `json:"raw,omitempty"`
	Host  []string   `json:"host,omitempty"`
	Port  string     `json:"port,omitempty"`
	Path  []string   `json:"path,omitempty"`
	Query []URLValue `json:"query,omitempty"`
}

type URLValue struct {
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Equals      bool   `json:"equals,omitempty"`
	Description string `json:"description,omitempty"`
}
