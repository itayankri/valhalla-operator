package resource

const valhallaDataPath = "/data"
const workerImage = "itayankri/valhalla-worker:latest"
const mapBuilderImage = "itayankri/valhalla-builder:latest"
const hirtoricalTrafficDataFetcherImage = "itayankri/predicted-traffic-fetcher:latest"
const liveTrafficDataFetcherImage = "itayankri/live-traffic-fetcher:latest"

const DeploymentSuffix = ""
const HorizontalPodAutoscalerSuffix = ""
const JobSuffix = "builder"
const CronJobSuffix = "predicted-traffic-fetcher"
const PersistentVolumeClaimSuffix = ""
const PodDisruptionBudgetSuffix = ""
const ServiceSuffix = ""
const containerPort = 8002
