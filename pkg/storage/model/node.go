package model

type Node struct {
	IP       string  `json:"ip"`
	Hostname string  `json:"hostname"`
	Port     string  `json:"port"`
	Drivers  Drivers `json:"drivers"`
}
