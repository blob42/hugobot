package main

import (
	"hugobot/export"
	"hugobot/feeds"
	"hugobot/static"
	"hugobot/utils"
	"errors"
	"log"
	"math"
	"path"
	"time"

	gum "git.sp4ke.com/sp4ke/gum.git"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/beeker1121/goque"
)

const (
	StaticDataExportInterval = 3600 * time.Second
	JobSchedulerInterval     = 60 * time.Second
	QDataDir                 = "./.data"
	MaxQueueJobs             = 100
	MaxSerialJob             = 100

	// Used in JobPool to avoid duplicate jobs for the same feed
	MapFeedJobFile = "map_feed_job"
)

var (
	SourcePriorityRange = Range{Min: 0, Max: 336 * time.Hour.Seconds()} // 2 weeks
	TargetPriorityRange = Range{Min: 0, Max: math.MaxUint8}

	SchedulerUpdates = make(chan *Job)
)

type Range struct {
	Min float64
	Max float64
}

func (r Range) Val() float64 {
	return r.Max - r.Min
}

// JobPool schedluer. Priodically schedule new jobs
type Scheduler struct {
	jobs       *JobPool
	jobUpdates chan *Job
	serialJobs chan *Job
}

func serialRun(inputJobs <-chan *Job) {
	for {
		j := <-inputJobs
		log.Printf("serial run %v", j)
		j.Handle()
	}
}

func (s *Scheduler) Run(m gum.UnitManager) {
	go serialRun(s.serialJobs) // These jobs run in series

	jobTimer := time.NewTicker(JobSchedulerInterval)
	staticExportTimer := time.NewTicker(StaticDataExportInterval)

	for {
		select {
		case <-jobTimer.C:
			log.Println("job heartbeat !")

			j, _ := s.jobs.Peek()
			if j != nil {
				log.Printf("peeking job: %s\n", j)
			}

			// If max pool jobs reached clean the pool
			if s.jobs.Length() >= MaxQueueJobs {
				s.jobs.Drop()
				s.panicAndShutdown(m)
				return
			}

			s.updateJobs()

			// Spawn job works
			s.dispatchJobs()

		case job := <-s.jobUpdates:
			log.Printf("job update recieved: %s", JobStatusMap[job.Status])

			switch job.Status {

			case JobStatusDone:
				log.Println("Job is done, removing from feedJobMap. New jobs for this feed can be added now.")

				// Remove job from feedJobMap
				err := s.jobs.DeleteMarkedJob(job)
				if err != nil {
					log.Fatal(err)
				}

				// Create export job for successful fetch jobs
				if job.JobType == JobTypeFetch {
					log.Println("Creating an export job")
					expJob, err := NewExportJob(job.Feed, 0)
					if err != nil {
						log.Fatal(err)
					}

					//log.Printf("export job: %+v", expJob)
					log.Printf("export job: %s\n", expJob)

					err = s.jobs.Enqueue(expJob)
					if err != nil {
						log.Fatal(err)
					}
				}

			case JobStatusFailed:
				//TODO: Store all failed jobs somewhere
				err := s.jobs.DeleteMarkedJob(job)
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("Job %s failed with error:  %s", job.Feed.Name, job.Err)
			}

		case <-staticExportTimer.C:
			log.Println("-------- export tick --------")

			log.Println("Exporting static data ...")
			err := static.HugoExportData()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Exporting weeks ...")
			err = export.ExportWeeks()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Exporting btc ...")
			err = export.ExportBTCAddresses()
			if err != nil {
				log.Fatal(err)
			}

		case <-m.ShouldStop():
			s.Shutdown()
			m.Done()

		}
	}
}

func (s *Scheduler) panicAndShutdown(m gum.UnitManager) {

	//err := s.jobs.Drop()
	//if err != nil {
	//log.Fatal(err)
	//}
	//TODO
	m.Panic(errors.New("max job queue exceeded"))

	s.Shutdown()

	m.Done()
}

func (s *Scheduler) Shutdown() {
	// Flush ongoing jobs back to job queue

	iter := s.jobs.feedJobMap.NewIterator(nil, nil)
	var markedDelete [][]byte
	for iter.Next() {
		key := iter.Key()
		log.Printf("Putting job %s back to queue", key)
		value := iter.Value()
		//log.Println("value ", value)
		job, err := JobFromBytes(value)
		if err != nil {
			log.Fatal(err)
		}

		err = s.jobs.Enqueue(job)
		if err != nil {
			log.Fatal(err)
		}

		markedDelete = append(markedDelete, key)

	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Fatal(err)
	}

	for _, k := range markedDelete {

		err := s.jobs.feedJobMap.Delete(k, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Close jobpool queue
	err = s.jobs.Close()
	if err != nil {
		panic(err)
	}

}

// Dispatchs all jobs in the job pool to task workers
func (s *Scheduler) dispatchJobs() {
	//log.Println("dispatching ...")
	jobsLength := int(s.jobs.Length())
	for i := 0; i < jobsLength; i++ {
		log.Printf("Dequeing %d/%d", i, s.jobs.Length())

		j, err := s.jobs.Dequeue()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Dispatching  ", j)

		if j.Serial {
			s.serialJobs <- j
		} else {
			go j.Handle()
		}
	}

}

// Gets all available feeds and creates
// a new Job if time.Now() - feed.last_refresh >= feed.interval
func (s *Scheduler) updateJobs() {
	//
	// Get all feeds
	//
	// For each feed compare Now() vs last refresh
	// If now() - last_refresh >= refresh interval -> create job

	feeds, err := feeds.ListFeeds()
	if err != nil {
		log.Fatal(err)
	}

	//log.Printf("updating jobs for %d feeds\n", len(feeds))

	// Check all jobs
	for _, f := range feeds {
		//log.Printf("checking feed %s: %s\n", f.Name, f.Url)
		//log.Printf("Seconds since last refresh %f", time.Since(f.LastRefresh.Time).Seconds())
		//log.Printf("Refresh interval %f", f.Interval)

		if delta, ok := f.ShouldRefresh(); ok {
			log.Printf("Refreshing %s -- %f seconds since last.", f.Name, delta)

			// If there is already a job with this feed id skip and return an empty job
			feedId := utils.IntToBytes(f.FeedID)

			_, err := s.jobs.feedJobMap.Get(feedId, nil)
			if err == nil {
				log.Println("Job already exists for feed")
			} else if err != leveldb.ErrNotFound {
				log.Fatal(err)
			} else {

				// Priority is based on the delta time since last refresh
				// bigger delta == higher priority

				// Convert priority to smaller range priority
				// We use original range of `0 - 1 month` in seconds
				// Target range is uint8 `0 - 255`
				prio := MapPriority(delta)

				job, err := NewFetchJob(f, prio)
				if err != nil {
					panic(err)
				}

				err = s.jobs.Enqueue(job)
				if err != nil {
					panic(err)
				}

				err = s.jobs.MarkUniqJob(job)
				if err != nil {
					panic(err)
				}

			}

		}

	}

	jLen := s.jobs.Length()
	if jLen > 0 {
		log.Printf("jobs length = %+v\n", jLen)
	}
}

func NewScheduler() gum.WorkUnit {
	// Priority queue for jobs
	q, err := goque.OpenPriorityQueue(QDataDir, goque.DESC)
	if err != nil {
		panic(err)
	}

	// map[FEED_ID][JOB]
	// Used to avoid duplicate jobs in the queue for the same feed
	feedJobMapDB, err := leveldb.OpenFile(path.Join(QDataDir, MapFeedJobFile), nil)
	if err != nil {
		panic(err)
	}

	jobPool := &JobPool{
		Q:          q,
		maxJobs:    MaxQueueJobs,
		feedJobMap: feedJobMapDB,
	}

	// Restore all ongoing jobs in feedJobMap to the pool
	iter := feedJobMapDB.NewIterator(nil, nil)
	var markedDelete [][]byte
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		job, err := JobFromBytes(value)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Restoring uncomplete job %s back to job queue", job.ID)

		err = jobPool.Enqueue(job)
		if err != nil {
			log.Fatal(err)
		}

		markedDelete = append(markedDelete, key)
	}

	iter.Release()
	err = iter.Error()
	if err != nil {
		log.Fatal(err)
	}

	serialJobs := make(chan *Job, MaxSerialJob)

	return &Scheduler{

		jobs:       jobPool,
		jobUpdates: SchedulerUpdates,
		serialJobs: serialJobs,
	}
}

func NotifyScheduler(job *Job) {
	SchedulerUpdates <- job
}

func MapPriority(val float64) uint8 {
	newVal := (((val - SourcePriorityRange.Min) * TargetPriorityRange.Val()) /
		SourcePriorityRange.Val()) + TargetPriorityRange.Min

	return uint8(newVal)
}
