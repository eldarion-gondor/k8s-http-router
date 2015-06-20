# k8s-http-router

> Warning: this project is very new and might have some issues. We have been using it successfully in our testing clusters for a couple of months. Bug reports are welcome!g

This project has been pulled from the new Gondor backend built on Kubernetes. It simply routes HTTP requests to Kubernetes services using the HTTP Host header. Additional configuration can be specified, but nothing else is currently supported.

# Setup

To get k8s-http-router running on your Kubernetes cluster you can use the following replication controller:

    kind: ReplicationController
    apiVersion: v1beta3
    metadata:
        name: router
    spec:
        replicas: 3
        selector:
            name: router
        template:
            metadata:
              labels:
                name: router
            spec:
              containers:
                - name: router
                  image: <todo>
                  ports:
                    - name: router-http
                      containerPort: 80
                    - name: router-https
                      containerPort: 443

and service:

    kind: Service
    apiVersion: v1beta3
    metadata:
        name: router
        labels:
            name: router
    spec:
        selector:
            name: router
        ports:
            - name: router-http
              port: 80
            - name: router-https
              port: 443
        createExternalLoadBalancer: true

To add a service to route, use the `router` annotation:

    kind: Service
    apiVersion: v1beta3
    metadata:
        name: my-website
        labels:
            name: my-website
        annotations:
            router: '{"config": {}, "hosts": ["example.com"]}'
        spec:
            selector:
                name: my-website
            ports:
                - port: 80
                  targetPort: 8000

k8s-http-router will watch for service changes and keep its internal data structures up-to-date with changes in the Kubernetes service objects.

# Credit

Thanks to Google for the amazing [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes) project.

Thanks to the [flynn](https://github.com/flynn/flynn) project for a great router design this is inspired by, but stripped down and modified for Kubernetes.
