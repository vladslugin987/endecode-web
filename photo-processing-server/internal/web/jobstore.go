package web

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	redis "github.com/go-redis/redis/v8"
	"photo-processing-server/internal/config"
	"photo-processing-server/internal/services"
)

var (
	redisClient *redis.Client
	jobTTL      = 24 * time.Hour
	tokenTTL    = 24 * time.Hour
)

func InitJobStore(logger *services.Logger) {
	cfg := config.Load()
	// If Redis host is not set, skip
	if cfg.RedisHost == "" {
		logger.Log("JobStore: Redis not configured, using in-memory store")
		return
	}
	redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr()})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("JobStore: Redis unavailable, falling back to in-memory store")
		redisClient = nil
		return
	}
	logger.Log("JobStore: Redis enabled")
}

func jobKey(id string) string { return "job:" + id }
func tokenKey(token string) string { return "dl:" + token }
func lockKey(opPath string) string { return "lock:" + opPath }

func SaveJob(job *ProcessingJob) {
	jobs[job.ID] = job
	if redisClient != nil {
		ctx := context.Background()
		data, _ := json.Marshal(job)
		redisClient.Set(ctx, jobKey(job.ID), data, jobTTL)
	}
}

func GetJob(id string) (*ProcessingJob, bool) {
	if redisClient != nil {
		ctx := context.Background()
		s, err := redisClient.Get(ctx, jobKey(id)).Result()
		if err == nil && s != "" {
			var j ProcessingJob
			if json.Unmarshal([]byte(s), &j) == nil {
				return &j, true
			}
		}
	}
	j, ok := jobs[id]
	return j, ok
}

func UpdateJobProgress(id string, progress float64) {
	if j, ok := jobs[id]; ok {
		j.Progress = progress
	}
	if redisClient != nil {
		if j, ok := GetJob(id); ok {
			j.Progress = progress
			SaveJob(j)
		}
	}
}

func SetJobStatus(id string, status string, errMsg string) {
	if j, ok := jobs[id]; ok {
		j.Status = status
		j.Error = errMsg
		if status == "completed" {
			j.Progress = 1.0
		}
	}
	if redisClient != nil {
		if j, ok := GetJob(id); ok {
			j.Status = status
			j.Error = errMsg
			if status == "completed" {
				j.Progress = 1.0
			}
			SaveJob(j)
		}
	}
}

func SetJobResult(id string, result interface{}) {
	if j, ok := jobs[id]; ok {
		j.Result = result
	}
	if redisClient != nil {
		if j, ok := GetJob(id); ok {
			j.Result = result
			SaveJob(j)
		}
	}
}

// Download token storage
func SaveDownloadToken(token string, path string) {
	downloadTokens[token] = path
	if redisClient != nil {
		ctx := context.Background()
		redisClient.Set(ctx, tokenKey(token), path, tokenTTL)
	}
}

func GetDownloadPath(token string) (string, bool) {
	if redisClient != nil {
		ctx := context.Background()
		s, err := redisClient.Get(ctx, tokenKey(token)).Result()
		if err == nil && s != "" {
			return s, true
		}
	}
	p, ok := downloadTokens[token]
	return p, ok
}

func DeleteDownloadToken(token string) {
	delete(downloadTokens, token)
	if redisClient != nil {
		ctx := context.Background()
		redisClient.Del(ctx, tokenKey(token))
	}
}

// Distributed operation locks
// TryAcquireOpLock tries to acquire a lock for opPath storing jobId, returns existing jobId if busy
func TryAcquireOpLock(opPath string, jobId string, ttl time.Duration) (acquired bool, existingJobId string) {
	if redisClient != nil {
		ctx := context.Background()
		// If lock exists, return its value (jobId)
		if val, err := redisClient.Get(ctx, lockKey(opPath)).Result(); err == nil && val != "" {
			return false, val
		}
		// Try to set if not exists
		ok, err := redisClient.SetNX(ctx, lockKey(opPath), jobId, ttl).Result()
		if err == nil && ok {
			return true, ""
		}
		// If failed, read again
		if val, err := redisClient.Get(ctx, lockKey(opPath)).Result(); err == nil && val != "" {
			return false, val
		}
		return false, ""
	}
	// Fallback to in-memory
	activeMutex.Lock()
	defer activeMutex.Unlock()
	if existing, ok := activeOps[opPath]; ok {
		return false, existing
	}
	activeOps[opPath] = jobId
	return true, ""
}

// ReleaseOpLock releases lock only if value matches jobId
func ReleaseOpLock(opPath string, jobId string) {
	if redisClient != nil {
		ctx := context.Background()
		// simple check-and-del (not atomic across processes but acceptable here)
		if val, err := redisClient.Get(ctx, lockKey(opPath)).Result(); err == nil && val == jobId {
			redisClient.Del(ctx, lockKey(opPath))
		}
		return
	}
	activeMutex.Lock()
	defer activeMutex.Unlock()
	if val, ok := activeOps[opPath]; ok && val == jobId {
		delete(activeOps, opPath)
	}
}

// Admin utilities

type JobSummary struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	StartTime time.Time `json:"startTime"`
}

func ListJobs() []JobSummary {
	items := make([]JobSummary, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, JobSummary{
			ID:        j.ID,
			Status:    j.Status,
			Progress:  j.Progress,
			StartTime: j.StartTime,
		})
	}
	// Sort by StartTime desc
	sort.Slice(items, func(i, j int) bool {
		return items[i].StartTime.After(items[j].StartTime)
	})
	return items
} 