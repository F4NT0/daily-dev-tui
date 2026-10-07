package main

type Field struct {
	Key   string
	Label string
	Path  bool // path parameter (otherwise query)
	Req   bool
}

type Endpoint struct {
	Group  string
	Name   string
	Path   string
	Fields []Field
}

var (
	fLimit  = Field{Key: "limit", Label: "limit (1-50)"}
	fCursor = Field{Key: "cursor", Label: "cursor (pagination)"}
	fQ      = Field{Key: "q", Label: "search term", Req: true}
	fTime   = Field{Key: "time", Label: "time (day, week, month, year, all)"}
)

func pathF(k, l string) Field { return Field{Key: k, Label: l, Path: true, Req: true} }

var endpoints = []Endpoint{
	{"Feeds", "For You", "/feeds/foryou", []Field{fLimit, fCursor}},
	{"Feeds", "Popular", "/feeds/popular", []Field{{Key: "tags", Label: "tags (comma separated)"}, fLimit, fCursor}},
	{"Feeds", "Discussed", "/feeds/discussed", []Field{{Key: "period", Label: "period in days (1-30)"}, {Key: "tag", Label: "tag"}, {Key: "source", Label: "source id"}, fLimit, fCursor}},
	{"Feeds", "By tag", "/feeds/tag/{tag}", []Field{pathF("tag", "tag"), fLimit, fCursor}},
	{"Feeds", "By source", "/feeds/source/{source}", []Field{pathF("source", "source id or handle"), fLimit, fCursor}},
	{"Posts", "Post details", "/posts/{id}", []Field{pathF("id", "post id")}},
	{"Posts", "Post comments", "/posts/{id}/comments", []Field{pathF("id", "post id"), {Key: "sort", Label: "sort (oldest/newest)"}, fLimit, fCursor}},
	{"Search", "Search posts", "/search/posts", []Field{fQ, fTime, fLimit, fCursor}},
	{"Search", "Search tags", "/search/tags", []Field{fQ}},
	{"Search", "Search sources", "/search/sources", []Field{fQ}},
	{"Bookmarks", "My bookmarks", "/bookmarks/", []Field{{Key: "unreadOnly", Label: "unreadOnly (true/false)"}, {Key: "listId", Label: "list id"}, fLimit, fCursor}},
	{"Bookmarks", "Search bookmarks", "/bookmarks/search", []Field{fQ, fLimit, fCursor}},
	{"Bookmarks", "Bookmark lists", "/bookmarks/lists", nil},
	{"Custom Feeds", "List custom feeds", "/feeds/custom/", nil},
	{"Custom Feeds", "Custom feed posts", "/feeds/custom/{feedId}", []Field{pathF("feedId", "feed id"), fLimit, fCursor}},
	{"Custom Feeds", "Custom feed info", "/feeds/custom/{feedId}/info", []Field{pathF("feedId", "feed id")}},
	{"Custom Feeds", "Advanced settings", "/feeds/custom/advanced-settings", nil},
	{"Feed Filters", "Global feed filters", "/feeds/filters/", nil},
	{"Notifications", "Notifications", "/notifications/", []Field{fLimit, fCursor}},
	{"Notifications", "Unread count", "/notifications/unread/count", nil},
	{"Profile", "My profile", "/profile/", nil},
	{"Profile", "Search stack", "/profile/stack/search", []Field{fQ}},
	{"Profile", "My stack", "/profile/stack/", nil},
	{"Profile", "My experiences", "/profile/experiences/", []Field{{Key: "type", Label: "type (work, education, project, ...)"}}},
	{"Profile", "Experience by id", "/profile/experiences/{id}", []Field{pathF("id", "experience id")}},
	{"Tags", "All tags", "/tags/", nil},
	{"Recommend", "By keyword (experimental)", "/recommend/keyword", []Field{fQ, fLimit}},
}
