package models

type LatestBuild struct {
	ID             uint16  `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	BuildID        uint32  `json:"build_id"`
	RepositoryName *string `json:"repository_name" gorm:"index:repo"`
}
