// Copyright by Danyil Dobryvechir 2020 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)
// mode: test (no changes), fix, compilation
// compression: 0 -no, 1 - to file,

package main


const (
	copyright = "Copyright by Danyil Dobryvechir 2020 Christmas"
)

type JsHelper struct {
	Mode          string   `json:"mode"`
	Src           []string `json:"src"`
	AllSrc        []string `json:"allSrc"`
	SrcMask       string   `json:"srcMask"`
	VersionFile   string   `json:"versionFile"`
	VersionSearch string   `json:"versionSearch"`
}

func main() {

	//searchOptions := dvsearch.GenerateSearchOptions(search, replace, options, pattern)
	//if len(searchOptions.Search) == 0 {
	//	fmt.Println("Search cannot be empty")
	//	return
	//}
	//all, found := dvsearch.SearchProcessDir(startDir, searchOptions)

}
