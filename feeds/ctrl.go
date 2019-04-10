package feeds

import (
	"git.sp4ke.com/sp4ke/hugobot/v3/types"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	MsgOK = "OK"
)

var (
	ErrNotInt = "expected int"
)

type FeedCtrl struct{}

func (ctrl FeedCtrl) Create(c *gin.Context) {

	var feedForm FeedForm
	feedModel := new(Feed)

	if err := c.ShouldBindJSON(&feedForm); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"status":  http.StatusNotAcceptable,
			"message": "invalid form",
			"form":    feedForm})
		c.Abort()
		return
	}

	feedModel.Name = feedForm.Name
	feedModel.Url = feedForm.Url
	feedModel.Format = feedForm.Format
	feedModel.Section = feedForm.Section
	feedModel.Categories = types.StringList(feedForm.Categories)

	err := feedModel.Write()

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotAcceptable,
			gin.H{"status": http.StatusNotAcceptable, "error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": MsgOK})
}

func (ctrl FeedCtrl) List(c *gin.Context) {

	feeds, err := ListFeeds()
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"error":  err.Error(),
			"status": http.StatusNotAcceptable,
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "result": feeds})

}

func (ctrl FeedCtrl) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"error":  ErrNotInt,
			"status": http.StatusNotAcceptable,
		})
		c.Abort()
		return
	}
	err = DeleteById(id)

	sqlErr, isSqlErr := err.(sqlite3.Error)
	if err != nil {

		if isSqlErr {

			c.JSON(http.StatusInternalServerError,
				gin.H{
					"error":  sqlErr.Error(),
					"status": http.StatusInternalServerError,
				})

		} else {

			var status int

			switch err {
			case ErrDoesNotExist:
				status = http.StatusNotFound
			default:
				status = http.StatusInternalServerError
			}

			c.JSON(status,
				gin.H{"error": err.Error(), "status": status})

		}

		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": MsgOK})
}
