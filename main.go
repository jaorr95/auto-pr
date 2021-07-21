package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	BaseUrl = "https://api.bitbucket.org/2.0"
)

func main()  {
	log.Println("Starting...")
	var (
		username     string
		password     string
		workspace    string
		repoSlug     string
		sourceBranch string
		targetBranch string
		title        string
	)

	flag.StringVar(&username, "username", "", "Nombre de usuario de la cuenta de bitbucket. Obligatorio")
	flag.StringVar(&password, "password", "", "Contrasena de la cuenta de bitbucket. Obligatorio")
	flag.StringVar(&workspace, "workspace", "", "Workspace donde se encuentra el repositorio de bitbucket. Obligatorio")
	flag.StringVar(&repoSlug, "repo", "", "Nombre del repositorio. Obligatorio")
	flag.StringVar(&sourceBranch, "src-branch", "", "Rama origen del pull request. Obligatorio")
	flag.StringVar(&targetBranch, "target-branch", "master", "Rama destino del pull request")
	flag.StringVar(&title, "title-pr", "Creado con auto-pr", "Titulo del pull request")

	flag.Parse()
	ValidateFlagString("Faltan parametros obligatorios" ,username, password, workspace, repoSlug, sourceBranch)

	createPullRequestResponse := CreatePullRequest(username, password, workspace, repoSlug, sourceBranch, targetBranch, title)
	log.Println("Pull Request creado con exito")
	// Convierto de manera explicita cada elemento dentro del interface
	mergeUrl := createPullRequestResponse["links"].(map[string]interface{})["merge"].(map[string]interface{})["href"].(string)
	log.Printf("Url para hacer merge del pull request %s\n", mergeUrl)

	MergePullRequest(username, password, mergeUrl)
	log.Println("Pull request mergeado con exito")

}

func CreatePullRequest(username, password, workspace, repoSlug, sourceBranch, targetBranch, title string) map[string]interface{}{
	clientHttp := &http.Client{}
	pathFormat := "repositories/{workspace}/{repo_slug}/pullrequests"
	path := strings.Replace(pathFormat, "{workspace}", workspace, -1)
	path = strings.Replace(path, "{repo_slug}", repoSlug, -1)
	path = fmt.Sprintf("%s/%s", BaseUrl, path)
	log.Printf("Haciendo peticion a la url %s\n", path)

	requestBody := map[string]interface{} {
		"title":               title,
		"close_source_branch": true,
		"source": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": sourceBranch,
			},
		},
		"destination": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": targetBranch,
			},
		},
	}

	buf, _ := json.Marshal(requestBody)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(buf))
	CheckErr(err, "Ocurrio un error en la creacion de la peticion para crear el pull request!")

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := clientHttp.Do(req)
	CheckErr(err, "Ocurrio un error durante la peticion para crear el pull request!")

	body := DecodeResponseRequest(http.StatusCreated, *resp)
	return body
	
}


func MergePullRequest(username, password, mergeUrl string) map[string]interface{} {

	log.Println("Haciendo merge del pull request...")

	clientHttp := &http.Client{}

	requestBody := map[string]interface{} {

		"close_source_branch": true,
		"merge_strategy": "merge_commit",
	}
	buf, _ := json.Marshal(requestBody)

	req, err := http.NewRequest(http.MethodPost, mergeUrl, bytes.NewBuffer(buf))
	CheckErr(err, fmt.Sprintf("Ocurrio un error un error en la creacion de la peticion para hacer merge del pull request en la url %s", mergeUrl))
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")


	resp, err := clientHttp.Do(req)
	CheckErr(err, fmt.Sprintf("Ocurrio un error en la peticion para hacer merge del pull request en la url %s", mergeUrl))

	body := DecodeResponseRequest(http.StatusOK, *resp)
	return body

}

func DecodeResponseRequest(expectedStatus int, response http.Response) map[string]interface{} {

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	CheckErr(err, "Ocurrio un error leyendo la respuesta de la peticion!")

	contentBody := make(map[string]interface{})

	if str := string(body); str != "" {
		err = json.Unmarshal(body, &contentBody)
		CheckErr(err, "Ocurrio un error encodeando el cuerpo de la peticion a json!")
	}


	if response.StatusCode != expectedStatus {
		log.Printf("EL estatus devuelto no fue el esperado, se obtuvo %d, se esperaba %d", response.StatusCode, expectedStatus)
		log.Fatal(string(body))
	}

	return contentBody

}

func ValidateFlagString(msg string, values...string) {
	for _, value := range values {
		if value == "" {
			log.Fatalln(msg)
		}
	}
}

func CheckErr(err error, msg string) {
	if err != nil {
		log.Println(msg)
		log.Fatal(err)
	}
}




