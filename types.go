package main

type configRoot struct {
	Tenants []tenantConfigStruct `json:"tenants"`
}

type tenantConfigStruct struct {
	ID    string             `json:"id"`
	Teams []teamConfigStruct `json:"teams"`
}

type teamConfigStruct struct {
	ID     string            `json:"id"`
	Filter string            `json:"filter"`
	Tags   []tagConfigStruct `json:"tags"`
}

type tagConfigStruct struct {
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}
