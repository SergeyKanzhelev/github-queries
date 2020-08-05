# Query some GitHub statistics

Running some GitHub statistics periodically.

## Set up

Configure pull-requests@apmtips.iam.gserviceaccount.com as an Editor for the spreadsheet.

## Run program once

kubectl create job --from=cronjob/sig-node-prs sig-node-prs-manual
kubectl get jobs
kubectl delete job sig-node-prs-manual