package status

import "gopro/core/component"

type Response struct {
	Version            Version                 `json:"version"`
	Players            Players                 `json:"players"`
	Description        component.TextComponent `json:"description"`
	Favicon            string                  `json:"favicon"`
	EnforcesSecureChat bool                    `json:"enforcesSecureChat"`
	PreviewsChat       bool                    `json:"previewsChat"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type Players struct {
	Max    int            `json:"max"`
	Online int            `json:"online"`
	Sample []SamplePlauer `json:"sample"`
}

type SamplePlauer struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}
