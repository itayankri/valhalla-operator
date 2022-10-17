package resource

const valhallaDataPath = "/data"
const workerImage = "itayankri/valhalla-worker:latest"
const mapBuilderImage = "itayankri/valhalla-builder:latest"
const hirtoricalTrafficDataFetcherImage = "itayankri/valhalla-historical-fetcher:latest"
const liveTrafficDataFetcherImage = "itayankri/valhalla-live-fetcher:latest"

const DeploymentSuffix = ""
const HorizontalPodAutoscalerSuffix = ""
const JobSuffix = "builder"
const CronJobSuffix = "historical-traffc-data-fetcher"
const PersistentVolumeClaimSuffix = ""
const PodDisruptionBudgetSuffix = ""
const ServiceSuffix = ""
const containerPort = 8002
