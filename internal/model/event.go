package model

type Event struct {
	Type    string `json:"type"`
	Service string `json:"service"`
	Env     string `json:"env"`
	Version string `json:"version"`
}
