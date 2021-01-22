package lunchmoney

// TagsResponse is the response from getting all tags.
type TagsResponse []*Tag

// Tag is a single LM tag.
type Tag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
