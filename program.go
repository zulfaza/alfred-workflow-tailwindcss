package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"

	aw "github.com/deanishe/awgo"
)

type AlgoliaData struct {
	AppId  string
	Index  string
	APIKey string
	Limit  int
}

var (
	dataEnv *AlgoliaData
	query   string

	wf *aw.Workflow
)

func init() {

	wf = aw.New()
}

func formatHierarchy(data map[string]interface{}) (string, string) {
	var result string
	keys := make([]string, 0, len(data))
	for key, value := range data {
		if value != nil {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	isFirst := true
	for _, k := range keys {
		value := data[k]
		if value != nil {
			valueStr := value.(string)
			if isFirst {
				result = valueStr
			} else {
				result = fmt.Sprintf("%v > %v", result, valueStr)
			}
		}
		isFirst = false
	}
	return result, keys[len(keys)-1]
}

func formatResult(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	jsonStr, _ := json.Marshal(data)
	log.Println(string(jsonStr))
	for _, keyCheck := range []string{"objectID", "hierarchy", "content", "url", "anchor"} {
		if keyCheck == "hierarchy" {
			mapedData := data["hierarchy"].(map[string]interface{})
			var lastLevel string
			result["subtitle"], lastLevel = formatHierarchy(mapedData)
			if value, ok := data["type"]; ok && value != nil && strings.HasPrefix(value.(string), "lv") {
				result["title"] = mapedData[value.(string)].(string)
			} else {
				result["title"] = mapedData[lastLevel].(string)
			}
		} else if value, ok := data[keyCheck]; ok && value != nil {
			result[keyCheck] = value.(string)
		}
	}

	return result
}

func algolia() {
	wf.Args()
	flag.Parse()

	if args := flag.Args(); len(args) > 0 {
		query = args[0]
	}

	icon := &aw.Icon{Value: "./icon.png", Type: aw.IconTypeImage}

	if query != "" {
		wf.Filter(query)
	} else {
		wf.NewWarningItem("No keyword", "please type any keyword").Icon(icon)
		wf.SendFeedback()
		return
	}

	dataEnv = &AlgoliaData{}
	if err := wf.Config.To(dataEnv); err != nil {
		panic(err)
	}
	client := search.NewClient(dataEnv.AppId, dataEnv.APIKey)

	index := client.InitIndex(dataEnv.Index)

	params := []interface{}{
		opt.FacetFilter("version:v3"),
		opt.HitsPerPage(dataEnv.Limit),
	}
	res, err := index.Search(query, params...)

	if err != nil || len(res.Hits) < 1 {
		wf.NewWarningItem("No matching items", "Try a different query?").Icon(icon)
	}

	for _, v := range res.Hits {
		result := formatResult(v)
		wf.NewItem(result["title"]).
			Subtitle(result["subtitle"]).
			UID(result["objectID"]).
			Valid(true).
			Arg(result["url"]).
			Copytext(result["url"]).
			Quicklook(result["url"]).
			Icon(icon)
	}

	wf.SendFeedback()

}

/**
Aplication Id : KNPXZI5B0M
aplication key : 5fc87cef58bb80203d2207578309fab6
index name : tailwindcss
**/
func main() {

	wf.Run(algolia)
}
