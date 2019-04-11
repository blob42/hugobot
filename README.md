**MIRRORED FROM**: https://git.sp4ke.com/sp4ke/hugobot

# HUGOBOT

*hugobot* is a an automated content fetch and aggregation bot for [Hugo][hugo] data
driven websites. It has the following features:


## Data fetch

- Add feeds to the bot in the `feeds` sqlite table
- Currently handles these types of feeds: `RSS`, `Github Releases`, `Newsletters`
- Define your own feed types by implementing the `JobHandler` interface (see
  `handlers/handlers.go`).
- Hugobot automatically fetch new posts from the feeds you defined
- It runs periodically to download new posts in the defined feeds.
- Storage is done with sqlite. 
- The scheduler can handle any number of tasks and uses leveldb for
  caching/resuming jobs.


## Hugo export

- Data is automatically exported to the configured Hugo website path.
- It can export `markdown` files or `json/toml` data files
- All fields in the exported files can be customized
- You can define custom output formats by using the `FormatHandler` interface.


## API

- *hugobot* also includes a webserver API that can be used with Hugo [Data
  Driven Mode][data-driven].

- WIP: Insert and query data 

- An example usage is the automated generation of Bitcoin addresses for new
  articles on [bitcointechweekly.com][btw-btc]

## Sqliteweb interface

- See Docker files


[data-driven]:https://gohugo.io/templates/data-templates/#data-driven-content
[btw-btc]:https://bitcointechweekly.com/btc/3Jv15g4G5LDnBJPDh1e2ja8NPnADzMxhVh
[hugo]:https://gohugo.io
