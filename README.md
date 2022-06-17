# indexs3client

[![Coverage Status](https://coveralls.io/repos/github/uc-cdis/indexs3client/badge.svg?branch=master)](https://coveralls.io/github/uc-cdis/indexs3client?branch=master)

S3 index client service

This container runs inside the ephemeral `indexing` pod, which is triggered by the `ssjdispatcher` pod based on data upload messages that sit on an AWS SQS queue.
