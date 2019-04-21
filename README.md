**MIRRORED FROM**: https://git.sp4ke.com/sp4ke/hugobot

# HUGOBOT

*hugobot* is a an automated content fetch and aggregation bot for [Hugo][hugo] data
driven websites. It has the following features:


## Data fetch

- Use the `feeds` table to  register feeds that will periodically get fetched, stored
  and exported into the hugo project.
- Currently handles these types of feeds: `RSS`, `Github Releases`, `Newsletters`
- Define your own feed types by implementing the `JobHandler` interface (see
  `handlers/handlers.go`).
- Hugobot automatically fetches new posts from the registered.
- Sqlite is used for storage. `feeds` and `posts` tables.
- The scheduler can handle any number of tasks and uses leveldb for
  caching/resuming jobs.


## Hugo export

- Data is automatically exported to the configured Hugo website path.
- It can export `markdown` files or `json/toml` data files.
- All fields in the exported files can be customized.
- You can define custom output formats by using the `FormatHandler` interface.
- You can register custom filters and post processing on exported posts to avoid 
changing the raw data stored in the db.
- You can force data export using the CLI.


## API

- Uses `gin-gonic`.

- *hugobot* also includes a webserver API that can be used with Hugo [Data
  Driven Mode][data-driven].

- Insert and query data from the db. This is still a WIP, you can easily 
  add the missing code on the API side to automate adding/querying data
  from the DB. 

- An example usage is the automated generation of Bitcoin addresses for new
  articles on [bitcointechweekly.com][btw-btc]

## Other

- Some commands are available through the CLI (`github.com/urfave/cli`), you
  can add your own custom commands.

## Sqliteweb interface

- See Docker files

## First time usage

- The database is automatically generated the first time you run the program.
  You can add your feeds straight into the sqlite db using your favorite sqlite GUI
  or the provided web gui in the docker-compose file.


[data-driven]:https://gohugo.io/templates/data-templates/#data-driven-content
[btw-btc]:https://bitcointechweekly.com/btc/3Jv15g4G5LDnBJPDh1e2ja8NPnADzMxhVh
[hugo]:https://gohugo.io
