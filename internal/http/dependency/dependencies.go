package dependency

import "template_sch/internal/platform/database"

type Dependencies struct {
	DB *database.Connections
}
