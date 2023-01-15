package cfa

type Request struct {
	Post interface{}            `json:"post"`
	Get  map[string]interface{} `json:"get"`
}

type YtbParams struct {
	DstDir     string `json:"dst_dir"`
	IsPlayList bool   `json:"is_playlist"`
	Format     string `json:"format"`
	Url        string `json:"url"`
}
