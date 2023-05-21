# Apacheweb-operator
**Apacheweb** operator provides basic features of *Apache HTTPD* server - web server and load balancer by using the extensions of Apache module *mod_proxy_balancer*

## Description
*Apache HTTPD* was the most popular HTTPD server in the near past and remains very popular in the Internet in nowadays, so the main goal of this operator is to bring *Apache HTTPD* features to *Kubernetes* world.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Run operator locally
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Deploy operator on the cluster manually
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/apache-operator:tag
```
	 
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/apache-operator:tag
```

### Create Apacheweb resource for load balancing
```yaml
apiVersion: apacheweb.arsenal.dev/v1alpha1
kind: Apacheweb
metadata:
  labels:
    app.kubernetes.io/name: apacheweb
    app.kubernetes.io/instance: apacheweb-sample
    app.kubernetes.io/part-of: apache-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: apache-operator
  name: apacheweb-sample
spec:
  serverName: "test.example.com"
  size: 2
  type: "lb" # type of operator: use "web" for web server or "lb" for load balancer
  loadBalancer:
    proto: https # proxy protocol
    path: /test # proxy path
    backEndService: remote-server # name of Service forwarding to
```

### Create Apacheweb resource for web server
```yaml
apiVersion: apacheweb.arsenal.dev/v1alpha1
kind: Apacheweb
metadata:
  labels:
    app.kubernetes.io/name: apacheweb
    app.kubernetes.io/instance: apacheweb-sample
    app.kubernetes.io/part-of: apache-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: apache-operator
  name: apacheweb-sample
spec:
  serverName: "test.example.com"
  size: 2
  type: "web" # type of operator: use "web" for web server or "lb" for load balancer
  webServer:
    documentRoot: /usr/local/apache2
    serverAdmin: test@example.com
    serverPort: 8888
```

```bash
kubectl apply -f apacheweb.yaml
```

If you use **Apacheweb** as load balancer, don't forget labeling the Servie resource which was put in *spec.loadBalancer.backEndService* - this service is used as a source of a target (remote) endpoints.

```bash
kubectl label service remote-server "kubernetes.io/service-name=remote-server"
```

```bash
k describe apachewebs.apacheweb.arsenal.dev apacheweb-sample
...
Status:
  End Points:
    Ip Address:  10.244.77.3
    Port:        80
    Proto:       http
    Status:      true
    Ip Address:  10.244.77.6
    Port:        80
    Proto:       http
    Status:      true
    Ip Address:  10.244.77.7
    Port:        80
    Proto:       http
    Status:      true
Events:
  Type    Reason   Age                    From                  Message
  ----    ------   ----                   ----                  -------
  Normal  Created  3m41s (x3 over 4m11s)  apacheweb-controller  EndPoint added IPAddress 10.244.77.3, port 80, protocol http, status true
  Normal  Created  3m41s (x2 over 3m43s)  apacheweb-controller  EndPoint added IPAddress 10.244.77.6, port 80, protocol http, status true
  Normal  Created  3m41s                  apacheweb-controller  EndPoint added IPAddress 10.244.77.7, port 80, protocol http, status true
```

```bash
k get pod -o wide|grep apacheweb-sample
apacheweb-sample-569996dcc9-8sxfv   1/1     Running   0          14m   10.244.77.10   k8s    <none>           <none>
apacheweb-sample-569996dcc9-plhcl   1/1     Running   0          14m   10.244.77.9    k8s    <none>           <none>

curl http://10.244.77.10:8080/test
hostname: remote-server-7fc9dffd6b-brqph

curl http://10.244.77.10:8080/test
hostname: remote-server-7fc9dffd6b-f29h6

curl http://10.244.77.10:8080/test
hostname: remote-server-7fc9dffd6b-x6lhb
```

You can use a *Service* for **Apacheweb** resources to destribute workloads between the load balancers

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
<<<<<<< HEAD
```

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

<<<<<<< HEAD
More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
=======
More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License
<<<<<<< HEAD

Copyright 2023 Oleksandr Krutko
=======
Copyright 2023.
>>>>>>> readme

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

>>>>>>> a2ed2d4adf5e40b999efe684e1844177f759ce07
=======
```
>>>>>>> readme
