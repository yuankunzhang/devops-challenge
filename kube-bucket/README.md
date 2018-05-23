# Kube-Bucket

Kube-Bucket defines a custom Kubernetes resource called "Bucket" and its related Kubernetes controller.

When a new Bucket resource is created, the controller creates a new S3 storage bucket. If you specify the `forceDelete` property in resource spec to be `true`, the target S3 bucket will be dropped once the resource is deleted.

## Dependency Management

It uses [dep](https://github.com/golang/dep) as dependency management tool.

## Building the Image

```shell
$ docker build -t kube-bucket:latest .
```

You should now have the `kube-bucket:latest` Docker image.

The multi-stages building technique is used in order to keep the image small.

## Applying the Bucket Resource Definition

The definition of the Bucket resource is in `crd/bucket-crd.yaml`. To create it, simply run:

```shell
$ kubectl create -f crd/bucket-crd.yaml
```

## Starting the Controller

### Using Docker Run

**Note: This approach is for testing purpose only.**

The controller can be started using `docker run`. Run the following command with all environment variables properly set.

```shell
$ docker run --rm -it --net host -v $HOME/.kube:/root/.kube:ro -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_REGION=$AWS_REGION  kube-bucket:latest
```

Now you should see logs directly from your terminal.

```
INFO[0000] Controller running...
```

### Using Kubernetes Deployment

I have already created a deployment manifest located in the `tests/` directory.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-bucket
spec:
  selector:
    matchLabels:
      app: kube-bucket
  replicas: 1
  template:
    metadata:
      labels:
        app: kube-bucket
    spec:
      containers:
      - name: kube-bucket
        image: kube-bucket:latest
        imagePullPolicy: Never
        env:
          - name: AWS_ACCESS_KEY_ID
            value: xxx
          - name: AWS_SECRET_ACCESS_KEY
            value: xxx
          - name: AWS_REGION
            value: xxx
```

Modify the `env` field accordingly, then run:

```shell
$ kubectl create -f kube-bucket-deployment.yaml
```

After the pod starts up, you can via its logs using the following command.

```shell
$ kubectl logs -f $(kubectl get po -l app=kube-bucket -o=name)
time="2018-05-23T22:35:40Z" level=info msg="Controller running..."
```

## Creating a Bucket

To create a Bucket resource, define a `test-bucket.yaml` file with below contents.

```yaml
apiVersion: storagek8s.honestbee.io/v1
kind: Bucket
metadata:
  name: my-bucket
spec:
  bucketName: me.yuankun.my-awesome-bucket
  region: ap-southeast-1
  forceDelete: true
```

Then run `kubectl create -f test-bucket.yaml`. You shall now see the following output from the controller's side:

```shell
INFO[0400] creating resource: default/my-bucket
INFO[0400] checking if bucket me.yuankun.my-awesome-bucket exists
INFO[0401] creating bucket me.yuankun.my-awesome-bucket
INFO[0403] creating bucket me.yuankun.my-awesome-bucket completed
```

And when you run `kubectl delete bucket my-bucket`, you shall see the following output.

```shell
INFO[0456] deleting resource: default/my-bucket
INFO[0456] deleting bucket me.yuankun.my-awesome-bucket
INFO[0457] 0 objects deleted from bucket me.yuankun.my-awesome-bucket
INFO[0457] deleting bucket me.yuankun.my-awesome-bucket completed
```

## Follow ups

To be done:

- The "update" event is not implemented yet. When the Bucket resource changes its bucket name, the controller needs to delete the original bucket and create a new one with the new bucket name.
- The `bucketExists()` function in `pkg/main/handler.go` seems buggy. Need more efforts on it.
- Using Secrets to distribute AWS credentials securely.
- Bucket name validation is good to have.
- Creating buckets backed by different cloud provider.
- How should other resources interact with the Bucket resource?