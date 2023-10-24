package main

type configRoot struct {
	Teams []teamConfigStruct `json:"teams"`
}

type teamConfigStruct struct {
	ID     string            `json:"id"`
	Filter string            `json:"filter"`
	Tags   []tagConfigStruct `json:"tags"`
}

type tagConfigStruct struct {
	DisplayName string   `json:"displayName"`
	Groups      []string `json:"groups"`
}
