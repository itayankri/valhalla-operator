package resource

const valhallaDataPath = "/data"
const workerImage = "itayankri/valhalla-worker:latest"
const mapBuilderImage = "itayankri/valhalla-builder:latest"

const DeploymentSuffix = ""
const HorizontalPodAutoscalerSuffix = ""
const JobSuffix = "builder"
const PersistentVolumeClaimSuffix = ""
const PodDisruptionBudgetSuffix = ""
const ServiceSuffix = ""
