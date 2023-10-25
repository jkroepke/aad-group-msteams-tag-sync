package main

type configRoot struct {
	Teams []teamConfigStruct `yaml:"teams"`
}

type teamConfigStruct struct {
	ID     string            `yaml:"id"`
	Filter string            `yaml:"filter"`
	Tags   []TagConfigStruct `yaml:"tags"`
}

type TagConfigStruct struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Groups      []string `yaml:"groups"`
}
