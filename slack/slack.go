package slack

type (
	Attachments struct {
		Attachments []Attachment `json:"attachments"`
	}

	Attachment struct {
		Fallback   string `json:"fallback"`
		AuthorName string `json:"author_name"`
		AuthorIcon string `json:"author_icon"`
		Pretext    string `json:"pretext"`
		Text       string `json:"text"`
		Title      string `json:"title"`
		TitleLink  string `json:"title_link"`
		ImageURL   string `json:"image_url"`
	}
)
