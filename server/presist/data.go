package presist

import (
	"context"
	"server/logger"
	"server/models"
	"time"
)

func AddJob(job models.Job) {
	ctx := context.Background()
	for _, image := range job.Images {

		// create hash for each image
		err := rds.HSet(ctx, job.Uid+"-"+image.Name, "name", image.Name, "path", image.Path).Err()
		if err != nil {
			logger.MyLog.Fatal(err)
		}

		// create list and push image hashes
		err = rds.RPush(ctx, "images:"+job.Uid, job.Uid+"-"+image.Name).Err()
		if err != nil {
			logger.MyLog.Fatal(err)
		}

	}

	// create hash and add list
	err := rds.HSet(ctx, "job:"+job.Uid, "uuid", job.Uid, "filter", job.Filter, "images", "images:"+job.Uid, "completed", "0", "started-processing", "0").Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func AddExpirationToJob(jobId string, timeToExpire time.Duration) {
	ctx := context.Background()
	err := rds.Expire(ctx, "job:"+jobId, timeToExpire).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func UpdateJobKey(jobId string, key string, value string) {
	ctx := context.Background()
	err := rds.HSet(ctx, "job:"+jobId, key, value).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func AddArchive(archiveId string) {
	ctx := context.Background()
	err := rds.SAdd(ctx, "archives-set", archiveId).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func RemoveArchive(archiveId string) {
	ctx := context.Background()
	err := rds.SRem(ctx, "archives-set", archiveId).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func AddUUID(uuid string) {
	ctx := context.Background()
	err := rds.SAdd(ctx, "uuid-set", uuid).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func GetAllJobs() []models.Job {
	ctx := context.Background()
	jobsResult, err := rds.Keys(ctx, "job:*").Result()
	if err != nil {
		logger.MyLog.Fatal(err)
	}

	if len(jobsResult) == 0 {
		return []models.Job{}
	}

	jobs := []models.Job{}
	for _, job := range jobsResult {
		completed, err := rds.HGet(ctx, job, "completed").Result()
		if err != nil {
			logger.MyLog.Fatal(err)
		}

		if completed == "0" {
			// generate job struct
			jobs = append(jobs, reconstructJob(job))
		}
	}

	return jobs
}

func reconstructJob(jobId string) models.Job {
	ctx := context.Background()
	result, err := rds.HGetAll(ctx, jobId).Result()

	if err != nil {
		panic(err)
	}

	job := models.Job{Filter: result["filter"], Uid: result["uuid"]}

	images, err := rds.LRange(ctx, result["images"], 0, -1).Result()
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(images); i++ {
		imageName, _ := rds.HGet(ctx, images[i], "name").Result()
		imagePath, _ := rds.HGet(ctx, images[i], "path").Result()
		job.Images = append(job.Images, models.Image{Name: imageName, Path: imagePath})
	}

	return job
}

func GetAllUUID() map[string]struct{} {
	ctx := context.Background()
	members, err := rds.SMembersMap(ctx, "uuid-set").Result()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
	return members
}

func GetAllArchives() map[string]struct{} {
	ctx := context.Background()
	members, err := rds.SMembersMap(ctx, "archives-set").Result()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
	return members
}

func RemoveUUID(uuid string) {
	ctx := context.Background()
	err := rds.SRem(ctx, "uuid-set", uuid).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}
}

func DeleteJob(jobId string) {
	ctx := context.Background()

	err := rds.Del(ctx, "job:"+jobId).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}

	// get all items in list then remove one by one then remove list
	images, err := rds.LRange(ctx, "images:"+jobId, 0, -1).Result()
	if err != nil {
		logger.MyLog.Fatal(err)
	}

	// delete each image
	for _, image := range images {
		err = rds.Del(ctx, image).Err()
		if err != nil {
			logger.MyLog.Fatal(err)
		}
	}

	// remove list
	err = rds.Del(ctx, "images:"+jobId).Err()
	if err != nil {
		logger.MyLog.Fatal(err)
	}

}
