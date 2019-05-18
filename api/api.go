// api package contains the main api endpoints which allows users to view resources
package api

import (
	"net/http"
	"fmt"
	"strconv"
	"vision/core/fileDriver"
	"vision/core/models"
	"strings"
	"io/ioutil"
	"encoding/json"
)

// configJson struct is the model for storing and holding the config data
var configJson models.ConfigModel

// aliases is a hashmap to create a one to one mapping
// between the alias name and resource path
var aliases map[string]string

// Store path to config file
var configJsonPath = "/etc/vision/config.json"

// Main function which loads config, creates alias hash and attaches handlers to routes
func Api() {
	// Read the config file and load it into memory as json
	// The connfig json will be stored in memory throughout the life
	// time of the object for fast retrieval of config
	loadConfigJson()

	// Create the alias map for fast retrieval.
	// The map will be stored in memory all the time
	// to access repeated file access
	createAliasMap()

	// Create route for / path and add handler function for it
	http.HandleFunc("/", apiHandler)

	// Create route for /aliases path and add handler function to it
	http.HandleFunc("/aliases", aliasHandler)
	http.ListenAndServe(":" + strconv.FormatInt(configJson.Port, 10), nil)
}

// This function will read config file and load the json into memory
func loadConfigJson() {
	file, _ := ioutil.ReadFile(configJsonPath)
	_ = json.Unmarshal([]byte(file), &configJson)
}

// This function will create a map (using the data in config json) to
// store aliases and their corresponding paths and store it in memory.
func createAliasMap() {
	aliasesTemp := make(map[string]string)
	for _, alias := range configJson.Aliases {
		aliasesTemp[alias.AliasName] = alias.AliasTo
	}
	aliases = aliasesTemp
}

// This is the handler for serving aliases. It returns the
// alias map as a list
func aliasHandler(w http.ResponseWriter, r *http.Request) {
	response := allAliases()
	fmt.Fprintf(w, response)
}

// This function creates a string representing the alias map with
// proper formatting.
func allAliases() (string) {
	aliasesSlice := make([]string, 0, 10)
	for key, value := range aliases {
		aliasesSlice = append(aliasesSlice, key + " : " + value)
	}

	if len(aliasesSlice) != 0 {
		aliasesString := strings.Join(aliasesSlice, "\n")
		return aliasesString
	}

	return ""
}

// This is the handler for root. It takes in a number of URL query params
// and it returns resource accordingly.
// It takes the followung URL params :
// 		path : The path to the resource to be viewed. For example it can be path to
//			   a log file
//		readFrom : It specifies the end from which the resource is to be read -
//				   head or tail. Accordingly it can take only two values: head|tail
//		limit : It denotes the number of lines to be read from that resource.
//				For example the number of lines of a log file to be read
//		posRegex : The value should be a regex. The regex will be used to filter lines
//				   from the resource
func apiHandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	pathSlice, isPath := r.URL.Query()["path"]
	readFromSlice, isReadFrom := r.URL.Query()["readFrom"]
	limitSlice, isLimit := r.URL.Query()["limit"]
	posRegexSlice, isPosRegex := r.URL.Query()["filterBy"]
	negRegexSlice, isNegRegex := r.URL.Query()["ignore"]
	aliasSlice, isAlias := r.URL.Query()["alias"]

	path, readFrom, limit, posRegex, negRegex, alias := "", "tail", int64(10), "", "", ""

	if isPath {
		path = pathSlice[0]
	}

	if isReadFrom {
		readFrom = readFromSlice[0]
	}

	if isLimit {
		limitTemp, err := strconv.ParseInt(limitSlice[0], 10, 64)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		limit = limitTemp
	}

	if isPosRegex {
		posRegex = posRegexSlice[0]
	}

	if isNegRegex {
		negRegex = negRegexSlice[0]
	}

	if isAlias {
		alias = aliasSlice[0]
	}

	request := &models.QueryHolder{
		Path: path,
		Alias: alias,
		ReadFrom: readFrom,
		Limit: limit,
		Regex: posRegex,
		NegateRegex: negRegex,
		Grep: "",
	}

	response, err := fileDriver.FileDriver(request, aliases, &configJson)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	fmt.Fprintf(w, response)
}
