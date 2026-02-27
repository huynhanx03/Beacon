package entity

type Embed struct {
	Title       string
	Description string
	Color       int
}

type Message struct {
	Content string
	Embeds  []Embed
}
