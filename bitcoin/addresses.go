package bitcoin

import (
	"hugobot/db"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	sqlite3 "github.com/mattn/go-sqlite3"
)

var DB = db.DB

const (
	DBBTCAddressesSchema = `CREATE TABLE IF NOT EXISTS btc_addresses (
		addr_id INTEGER PRIMARY KEY,
		address TEXT NOT NULL UNIQUE,
		address_position INTEGER NOT NULL DEFAULT 0,
		linked_article_title TEXT DEFAULT '',
		linked_article_id TEXT NOT NULL DEFAULT '',
		used INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	)`

	QueryUnusedAddress = `SELECT * FROM btc_addresses WHERE used = 0 LIMIT 1 `

	UpdateAddressQuery = `UPDATE btc_addresses 
		SET linked_article_id = ?,
		linked_article_title = ?,
		used = ?
		WHERE addr_id = ?
	`
)

type BTCAddress struct {
	ID                 int64  `db:"addr_id"`
	Address            string `db:"address"`
	AddrPosition       int64  `db:"address_position"`
	LinkedArticleTitle string `db:"linked_article_title"`
	LinkedArticleID    string `db:"linked_article_id"`
	Used               bool   `db:"used"`
	Synced             bool   `db:"synced"`
}

// TODO: Set address to synced
func (a *BTCAddress) SetSynced() error {
	a.Synced = true
	query := `UPDATE btc_addresses SET synced = :synced WHERE addr_id = :addr_id`
	_, err := DB.Handle.NamedExec(query, a)
	if err != nil {
		return err
	}

	return nil

}

func GetAddressByPos(pos int) (*BTCAddress, error) {
	var btcAddr BTCAddress
	err := DB.Handle.Get(&btcAddr,
		"SELECT * FROM btc_addresses WHERE address_position = ?",
		pos,
	)
	if err != nil {
		return nil, err
	}

	return &btcAddr, nil
}

func GetAddressByArticleID(artId string) (*BTCAddress, error) {
	var btcAddr BTCAddress
	err := DB.Handle.Get(&btcAddr,
		"SELECT * FROM btc_addresses WHERE linked_article_id = ?",
		artId,
	)
	if err != nil {
		return nil, err
	}

	return &btcAddr, nil
}

func GetAllUsedUnsyncedAddresses() ([]*BTCAddress, error) {
	var addrs []*BTCAddress
	err := DB.Handle.Select(&addrs,
		"SELECT * FROM btc_addresses WHERE used = 1 AND synced = 0",
	)
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

func GetNextUnused() (*BTCAddress, error) {
	var btcAddr BTCAddress
	err := DB.Handle.Get(&btcAddr, QueryUnusedAddress)
	if err != nil {
		return nil, err
	}
	return &btcAddr, nil
}

func GetAddressForArticle(artId string, artTitle string) (*BTCAddress, error) {
	// Check if article already has an assigned address
	addr, err := GetAddressByArticleID(artId)
	sqliteErr, isSqliteErr := err.(sqlite3.Error)

	if (isSqliteErr && sqliteErr.Code != sqlite3.ErrNotFound) ||
		(err != nil && !isSqliteErr && err != sql.ErrNoRows) {

		log.Println("err")
		return nil, err
	}

	if err == nil {
		// If different title update it
		if artTitle != addr.LinkedArticleTitle {
			addr.LinkedArticleTitle = artTitle
			// Store newly assigned address
			_, err = DB.Handle.Exec(UpdateAddressQuery,
				addr.LinkedArticleID,
				addr.LinkedArticleTitle,
				addr.Used,
				addr.ID,
			)
			if err != nil {
				return nil, err
			}
		}

		return addr, nil
	}

	// Get next unused address
	addr, err = GetNextUnused()
	if err != nil {
		return nil, err
	}

	addr.LinkedArticleID = artId
	addr.LinkedArticleTitle = artTitle
	addr.Used = true

	// Store newly assigned address
	_, err = DB.Handle.Exec(UpdateAddressQuery,
		addr.LinkedArticleID,
		addr.LinkedArticleTitle,
		addr.Used,
		addr.ID,
	)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func GetAddressCtrl(c *gin.Context) {
	artId := c.Query("articleId")
	artTitle := c.Query("articleTitle")

	addr, err := GetAddressForArticle(artId, artTitle)

	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"status": http.StatusBadRequest,
				"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"addr":   addr.Address,
	})

}

func init() {
	_, err := DB.Handle.Exec(DBBTCAddressesSchema)
	if err != nil {
		log.Fatal(err)
	}
}
