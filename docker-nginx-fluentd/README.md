# Docker-Nginx-Fluentd

This project deploys a Dockerized application serving a static website via Nginx. The request logs are sent to Fluentd, which in turn ships them to S3 or Elasticsearch.

The following official Docker images are used:

- [Nginx](https://hub.docker.com/_/nginx/)
- [Elasticsearch](https://hub.docker.com/_/elasticsearch/)
- [Kibana](https://hub.docker.com/_/kibana/)

The official [Fluentd](https://hub.docker.com/r/fluent/fluentd/) image does not provide support for Elasticsearch and S3. We use it as base image and build a new image upon it with `fluent-plugin-elasticsearch` and `fluent-plugin-s3` installed.

## Collecting Logs to Elasticsearch

**Note: Elasticsearch consumes a lot of RAM. On my Mac, I need to allocate at least 4GB of RAM to Docker machine in order to start the Elasticsearch container.**

```shell
$ docker-compose up -d --build
```

After the containers are created and started, you can now access the index page at http://localhost:8080.

The Kibana service can be see at http://localhost:5601. You need to set up the index name patter for Kibana by specifying `fluentd-*` to `Index name or pattern` and press the `Create` button. Then you can view Nginx request logs in Kibana.

The Fluentd service is configured by the file `fluentd/fluent.conf`. You can tune those configurations.

## Collecting Logs to S3

The first step is to create a S3 bucket.

Make sure you have AWS credentials properly set (either by [shared credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html) or by [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html)). And make sure you have [Terraform](https://www.terraform.io/) installed.

Go to the `terraform/` sub-directory and run `terraform apply`. After some seconds, a new S3 bucket will be created (if it hasn't been created yet).

The next step is to prepare the Fluentd config file.

```shell
# Create the config file.
$ cp fluentd/fluent.s3.conf.sample fluentd/fluent.s3.conf

# Edit this file by changing values of the following config items:
#   aws_key_id
#   aws_sec_key
#   s3_region
#   s3_bucket
```

Now let's start the services.

```shell
$ docker-compose -f docker-compose-s3.yml up -d --build
```

After the containers are created and started, you can now access the index page at http://localhost:8080. And the request logs are uploaded to the S3 bucket we've just created.

## Follow ups

To be done:

- Implement the deployment manifests for Kubernetes.
- It's not trivial to fine-tune the performance and resouce consuming of Elasticsearch, but that's beyond this task.

To dig deep:

- How to reload the configurations for Nginx/Fluentd/Elasticsearch/Kibana without restarting the containers?