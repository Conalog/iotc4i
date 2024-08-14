package iotc4i

// Field represents a field definition with a name and index range.
type Field struct {
	Name     string `json:"name,omitempty"` // Name is optional
	StartIdx int    `json:"startIdx"`
	EndIdx   int    `json:"endIdx"`
	Zerofill bool   `json:"zerofill,omitempty"` // Zerofill is optional
}
