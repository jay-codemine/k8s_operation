package cell

import (
	appv1alpha1 "gitee.com/jay-kim/appconfig-operator/api/v1alpha1"
	"time"
)

type AppConfigCell appv1alpha1.AppConfig

func (a *AppConfigCell) GetCreation() time.Time {
	return a.CreationTimestamp.Time
}

func (a *AppConfigCell) GetName() string {
	return a.Name
}
