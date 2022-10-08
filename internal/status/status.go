package status

const (
	AllProfilesReady         ValhallaConditionType = "AllProfilesReady"
	ClusterAvailable         ValhallaConditionType = "ClusterAvailable"
	ReconciliationSuccess    ValhallaConditionType = "ReconciliationSuccess"
	ReconciliationInProgress ValhallaConditionType = "ReconciliationInProgress"
)

type ValhallaConditionType string
