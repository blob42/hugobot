package feeds

type FeedForm struct {
	Name       string     `form:"name" binding:"required"`
	Url        string     `form:"url" binding:"required"`
	Format     FeedFormat `form:"format"`
	Categories []string   `form:"categories"`
	Section    string     `form:"section"`
}
