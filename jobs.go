package main

import (
	"hugobot/export"
	"hugobot/feeds"
	"hugobot/handlers"
	"hugobot/posts"
	"hugobot/utils"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/beeker1121/goque"
	"github.com/gofrs/uuid"
	"github.com/syndtr/goleveldb/leveldb"
)

type JobStatus int
type JobType int

const (
	JobStatusNew JobStatus = iota
	JobStatusQueued
	JobStatusDone
	JobStatusFailed
)

const (
	JobTypeFetch JobType = iota
	JobTypeExport
)

var (
	JobTypeMap = map[JobType]string{
		JobTypeFetch:  "fetch",
		JobTypeExport: "export",
	}

	JobStatusMap = map[JobStatus]string{
		JobStatusNew:    "new",
		JobStatusQueued: "queued",
		JobStatusDone:   "done",
		JobStatusFailed: "failed",
	}
)

func (js JobStatus) String() string {
	return JobStatusMap[js]
}

type Prioritizer interface {
	// Return job priority
	GetPriority() uint8
}

// Represents a Job to be done on a feed
// It could be any of: Poll, Fetch, Store
// Should implement Poller
type Job struct {
	ID     uuid.UUID
	Feed   *feeds.Feed
	Status JobStatus
	Data   []*posts.Post

	Priority uint8
	JobType  JobType
	Serial   bool // Should be run in a serial manner

	Err error

	Prioritizer
}

type Handler interface {
	Handle()
}

// GoRoutine method
func (job *Job) Handle() {
	var err error

	if job.JobType == JobTypeFetch {
		handler := handlers.GetFormatHandler(*job.Feed)
		err = handler.Handle(*job.Feed)
	} else if job.JobType == JobTypeExport {
		handler := export.NewHugoExporter()
		err = handler.Handle(*job.Feed)
	}

	if err != nil {
		job.Failed(err)
		return
	}
	//log.Println("Done for job type ", job.JobType)
	job.Done()
}

func (job *Job) Failed(err error) {
	errr := job.Feed.UpdateRefreshTime(time.Now())
	if errr != nil {
		log.Fatal(errr)
	}

	job.Status = JobStatusFailed
	job.Err = err
	NotifyScheduler(job)
}

func (job *Job) Done() {
	//TODO: only update refresh time after actual fetching
	//
	err := job.Feed.UpdateRefreshTime(time.Now())
	if err != nil {
		log.Fatal(err)
	}

	job.Status = JobStatusDone
	NotifyScheduler(job)
}

func (job *Job) GetPriority() uint8 {
	return job.Priority
}

func (job *Job) String() string {
	exp := map[string]interface{}{
		"jobId":    job.ID,
		"feed":     job.Feed.Name,
		"priority": job.Priority,
		"jobType":  JobTypeMap[job.JobType],
		"serial":   job.Serial,
		"err":      job.Err,
	}

	b, err := json.MarshalIndent(exp, "", " ")
	if err != nil {
		log.Printf("error printing job %s\n", err)
		return ""
	}
	return fmt.Sprintf(string(b))

}

// Decode object from []byte
func JobFromBytes(value []byte) (*Job, error) {
	buffer := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buffer)

	j := &Job{}

	err := dec.Decode(j)
	if err != nil {
		return nil, err
	}

	return j, nil
}

// helper function for jobs that accepts any
// value type, which is then encoded into a byte slice using
// encoding/gob.
func (job *Job) ToBytes() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(job); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func NewFetchJob(feed *feeds.Feed,
	priority uint8) (*Job, error) {

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:       uuid,
		Feed:     feed,
		Status:   JobStatusNew,
		JobType:  JobTypeFetch,
		Priority: priority,
		Serial:   feed.Serial,
	}

	return job, nil
}

func NewExportJob(feed *feeds.Feed,
	priority uint8) (*Job, error) {

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:       uuid,
		Feed:     feed,
		Status:   JobStatusNew,
		Priority: priority,
		JobType:  JobTypeExport,
	}

	return job, nil
}

type Queuer interface {
	Enqueue(job *Job) (*Job, error)
	Dequeue() (*Job, error)
	Close() error
	Drop() error // Clsoe and delete all jobs
	Length() uint64
	//Peek() (*Job, error)
	//PeekByID(id uint64) (*Job, error)

	// Returns item located at given offset starting from head
	// of queue without removing it
	//PeekByOffset(offset uint64) (*Job, error)
}

// Represents the queue of fetching todo jobs
type JobPool struct {
	// Actual jobs queue
	Q *goque.PriorityQueue

	// Handle queuing mechanics
	Queuer

	maxJobs int

	feedJobMap *leveldb.DB
}

func (jp *JobPool) Close() error {
	jp.Q.Close()

	err := jp.feedJobMap.Close()
	return err
}

func (jp *JobPool) Dequeue() (*Job, error) {
	item, err := jp.Q.Dequeue()
	if err != nil {
		return nil, err
	}
	j := &Job{}
	item.ToObject(j)

	//TODO: This is done when the job is done
	//feedId := utils.IntToBytes(j.Feed.ID)
	//err = jp.feedJobMap.Delete(feedId, nil)
	//if err != nil {
	//return nil, err
	//}

	return j, nil
}

func (jp *JobPool) DeleteMarkedJob(job *Job) error {
	var err error

	feedId := utils.IntToBytes(job.Feed.FeedID)
	err = jp.feedJobMap.Delete(feedId, nil)

	return err
}

// Mark a job in feedJobMap to avoid duplicates
func (jp *JobPool) MarkUniqJob(job *Job) error {

	// Mark the feed in the feedJobMap to avoid creating duplicates
	feedId := utils.IntToBytes(job.Feed.FeedID)

	jobData, err := job.ToBytes()
	if err != nil {
		return err
	}

	err = jp.feedJobMap.Put(feedId, jobData, nil)
	if err != nil {
		return err
	}

	return nil
}

func (jp *JobPool) Enqueue(job *Job) error {

	// Update job status
	job.Status = JobStatusQueued

	// Enqueue the job in the jobpool
	item, err := jp.Q.EnqueueObject(job.GetPriority(), job)
	if err != nil {
		return err
	}

	// Recode item to job
	j := &Job{}
	item.ToObject(j)

	return nil
}
func (jp *JobPool) Drop() {
	jp.Q.Drop()
}

func (jp *JobPool) Length() uint64 {
	return jp.Q.Length()
}

func (jp *JobPool) Peek() (*Job, error) {
	item, err := jp.Q.Peek()
	if err != nil {
		return nil, err
	}

	j := &Job{}
	item.ToObject(j)
	return j, err
}
