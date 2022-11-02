package resource

const valhallaDataPath = "/data"
const workerImage = "itayankri/valhalla-worker:latest"
const mapBuilderImage = "itayankri/valhalla-builder:latest"
const hirtoricalTrafficDataFetcherImage = "itayankri/valhalla-predicted-traffic:latest"
const liveTrafficDataFetcherImage = "itayankri/live-traffic-fetcher:latest"

const DeploymentSuffix = ""
const HorizontalPodAutoscalerSuffix = ""
const JobSuffix = "builder"
const CronJobSuffix = "predicted-traffic"
const PersistentVolumeClaimSuffix = ""
const PodDisruptionBudgetSuffix = ""
const ServiceSuffix = ""
const containerPort = 8002
