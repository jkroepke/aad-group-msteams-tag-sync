package main

type configRoot struct {
	Teams []teamConfigStruct `json:"teams"`
}

type teamConfigStruct struct {
	ID     string            `json:"id"`
	Filter string            `json:"filter"`
	Tags   []TagConfigStruct `json:"tags"`
}

type TagConfigStruct struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Groups      []string `json:"groups"`
}
