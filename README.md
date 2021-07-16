# indexs3client
S3 index client service

This container runs inside the ephemeral `indexing` pod, which is triggered by the `ssjdispatcher` pod based on data upload messages that sit on an AWS SQS queue.
