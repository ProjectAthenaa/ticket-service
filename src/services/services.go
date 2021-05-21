package services

import (
	"github.com/ProjectAthenaa/database-module"
	"os"
)

var DB, _ = database.Connect(os.Getenv("REDIS_URL"), os.Getenv("PG_URL"))
